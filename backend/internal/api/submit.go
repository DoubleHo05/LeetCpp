package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	maxFileBytes = 256 << 10 // 262144
	maxBodyBytes = 512 << 10 // 524288
	maxKeyLength = 128
)

var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// SubmitHandler serves POST /labs/{labId}/submissions per
// docs/contract/post-labs-labid-submissions.md. The evaluation order of the
// rules is the contract's error-precedence list; reordering either side alone
// is a contract break.
type SubmitHandler struct {
	store Store
	auth  Authenticator
	now   func() time.Time
}

// NewSubmitHandler builds the handler. now may be nil (defaults to time.Now).
func NewSubmitHandler(store Store, auth Authenticator, now func() time.Time) *SubmitHandler {
	if now == nil {
		now = time.Now
	}
	return &SubmitHandler{store: store, auth: auth, now: now}
}

// ServeHTTP implements http.Handler.
func (h *SubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, ok := h.auth.Authenticate(r)
	if !ok {
		writeProblem(w, r, http.StatusUnauthorized, "unauthenticated", "Unauthenticated",
			"sign in with a valid session token to submit work")
		return
	}

	labID := r.PathValue("labId")
	if !uuidRe.MatchString(labID) {
		writeProblem(w, r, http.StatusBadRequest, "validation-failed", "Validation Failed",
			"labId must be a UUID", FieldError{Field: "labId", Message: "must be a UUID (8-4-4-4-12 lowercase hex)"})
		return
	}

	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeProblem(w, r, http.StatusBadRequest, "invalid-idempotency-key", "Invalid Idempotency-Key",
			"Idempotency-Key header is required and must be 1-128 characters")
		return
	}
	if len(key) > maxKeyLength {
		writeProblem(w, r, http.StatusBadRequest, "invalid-idempotency-key", "Invalid Idempotency-Key",
			"Idempotency-Key must be at most 128 characters")
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		writeProblem(w, r, http.StatusUnsupportedMediaType, "unsupported-media-type", "Unsupported Media Type",
			"request body must be multipart/form-data")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := r.ParseMultipartForm(32 << 10); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeProblem(w, r, http.StatusRequestEntityTooLarge, "payload-too-large", "Payload Too Large",
				"request body exceeds the 512 KiB limit")
			return
		}
		writeProblem(w, r, http.StatusBadRequest, "validation-failed", "Validation Failed",
			"request body is not a valid multipart/form-data payload")
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "validation-failed", "Validation Failed",
			"file part is required", FieldError{Field: "file", Message: "required"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, maxFileBytes+1))
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, "validation-failed", "Validation Failed",
			"could not read the uploaded file")
		return
	}
	if len(content) > maxFileBytes {
		writeProblem(w, r, http.StatusRequestEntityTooLarge, "payload-too-large", "Payload Too Large",
			"file part exceeds the 256 KiB limit", FieldError{Field: "file", Message: "must be at most 262144 bytes"})
		return
	}

	if !allowedExtension(header.Filename) {
		writeProblem(w, r, http.StatusUnprocessableEntity, "unsupported-file-type", "Unsupported File Type",
			"filename extension must be one of .cpp .cc .cxx .h .hpp")
		return
	}

	fingerprint := fingerprint(header.Filename, content)
	if prevFP, prevSub, found := h.store.FindIdempotent(user.ID, labID, key); found {
		if prevFP == fingerprint {
			w.Header().Set("Idempotency-Replayed", "true")
			writeSubmission(w, http.StatusOK, prevSub)
			return
		}
		writeProblem(w, r, http.StatusBadRequest, "idempotency-key-reused", "Idempotency-Key Reused",
			"this Idempotency-Key was already used with a different payload")
		return
	}

	lab, found := h.store.Lab(labID)
	if !found || !user.MemberOf(lab.ClassID) {
		writeProblem(w, r, http.StatusNotFound, "lab-not-found", "Lab Not Found",
			"no lab with that id is visible to you")
		return
	}

	if user.Role != RoleStudent {
		writeProblem(w, r, http.StatusForbidden, "role-not-allowed", "Role Not Allowed",
			"only students may create submissions")
		return
	}

	now := h.now().UTC()
	if now.Before(lab.OpensAt) {
		writeProblem(w, r, http.StatusConflict, "submission-window-closed", "Submission Window Closed",
			"the submission window has not opened yet")
		return
	}
	if now.After(lab.ClosesAt) {
		writeProblem(w, r, http.StatusConflict, "submission-window-closed", "Submission Window Closed",
			"the submission window has closed")
		return
	}

	sum := sha256.Sum256(content)
	sub := &Submission{
		ID:        newID(),
		LabID:     labID,
		StudentID: user.ID,
		Filename:  header.Filename,
		SizeBytes: len(content),
		SHA256:    hex.EncodeToString(sum[:]),
		Status:    "queued",
		CreatedAt: now.Truncate(time.Second),
	}
	h.store.Create(sub, user.ID, labID, key, fingerprint)
	writeSubmission(w, http.StatusCreated, sub)
}

func writeSubmission(w http.ResponseWriter, status int, sub *Submission) {
	w.Header().Set("Location", "/submissions/"+sub.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(sub)
}

func allowedExtension(filename string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".cpp", ".cc", ".cxx", ".h", ".hpp":
		return true
	}
	return false
}

func fingerprint(filename string, content []byte) string {
	sum := sha256.Sum256(append(append([]byte(filename), 0x00), content...))
	return hex.EncodeToString(sum[:])
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
