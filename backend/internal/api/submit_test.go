package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	labID    = "7c9e6679-7425-40de-944b-e07fc1f90ae7"
	otherLab = "8d0e7780-8536-51ef-ab55-f19fa1dcb1bf"
	classA   = "class-a"
	classB   = "class-b"
)

type fixture struct {
	mux   *http.ServeMux
	store *InMemoryStore
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	store := NewInMemoryStore()
	now := time.Now().UTC()
	store.AddLab(&Lab{
		ID:       labID,
		ClassID:  classA,
		OpensAt:  now.Add(-time.Hour),
		ClosesAt: now.Add(time.Hour),
	})
	auth := BearerAuthenticator{Tokens: map[string]*User{
		"t-alice": {ID: "u-alice", Role: RoleStudent, ClassIDs: []string{classA}},
		"t-sam":   {ID: "u-sam", Role: RoleStudent, ClassIDs: []string{classA}},
		"t-bob":   {ID: "u-bob", Role: RoleStudent, ClassIDs: []string{classB}},
		"t-tina":  {ID: "u-tina", Role: RoleTA, ClassIDs: []string{classA}},
		"t-adam":  {ID: "u-adam", Role: RoleAdmin, ClassIDs: []string{classA}},
	}}
	mux := http.NewServeMux()
	mux.Handle("POST /labs/{labId}/submissions", NewSubmitHandler(store, auth, time.Now))
	return &fixture{mux: mux, store: store}
}

type reqOpts struct {
	token           string
	labID           string
	key             string
	omitKey         bool
	filename        string
	content         []byte
	rawBody         string
	contentType     string
	omitContentType bool
}

func (f *fixture) submit(t *testing.T, o reqOpts) *httptest.ResponseRecorder {
	t.Helper()
	if o.labID == "" {
		o.labID = labID
	}
	if o.key == "" && !o.omitKey {
		o.key = "key-1"
	}
	if o.filename == "" {
		o.filename = "sort.cpp"
	}
	if o.content == nil {
		o.content = []byte("#include <algorithm>\n")
	}

	var body io.Reader
	ct := o.contentType
	if o.rawBody != "" {
		body = strings.NewReader(o.rawBody)
		if ct == "" {
			ct = "application/json"
		}
	} else {
		buf := multipartBody(t, o.filename, o.content)
		body = buf
		if ct == "" {
			ct = buf.contentType
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/labs/"+o.labID+"/submissions", body)
	if o.token != "" {
		req.Header.Set("Authorization", "Bearer "+o.token)
	}
	if !o.omitKey {
		req.Header.Set("Idempotency-Key", o.key)
	}
	if !o.omitContentType {
		req.Header.Set("Content-Type", ct)
	}
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)
	return rec
}

type mpBuffer struct {
	*bytes.Buffer
	contentType string
}

func multipartBody(t *testing.T, filename string, content []byte) *mpBuffer {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return &mpBuffer{Buffer: &buf, contentType: mw.FormDataContentType()}
}

func decodeProblem(t *testing.T, rec *httptest.ResponseRecorder) Problem {
	t.Helper()
	var p Problem
	if err := json.NewDecoder(rec.Body).Decode(&p); err != nil {
		t.Fatalf("decode problem: %v (body %q)", err, rec.Body.String())
	}
	return p
}

func decodeSubmission(t *testing.T, rec *httptest.ResponseRecorder) Submission {
	t.Helper()
	var s Submission
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode submission: %v (body %q)", err, rec.Body.String())
	}
	return s
}

func wantProblem(t *testing.T, rec *httptest.ResponseRecorder, status int, typeSuffix string) Problem {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, status, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}
	p := decodeProblem(t, rec)
	if p.Type != problemBase+typeSuffix {
		t.Fatalf("type = %q, want %q", p.Type, problemBase+typeSuffix)
	}
	if p.Status != status {
		t.Fatalf("problem.status = %d, want %d", p.Status, status)
	}
	if p.Detail == "" {
		t.Fatal("problem.detail is empty")
	}
	return p
}

func wantCreated(t *testing.T, rec *httptest.ResponseRecorder, studentID, filename string, content []byte) Submission {
	t.Helper()
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %q)", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	s := decodeSubmission(t, rec)
	if s.ID == "" {
		t.Fatal("id is empty")
	}
	if s.LabID != labID {
		t.Fatalf("labId = %q, want %q", s.LabID, labID)
	}
	if s.StudentID != studentID {
		t.Fatalf("studentId = %q, want %q", s.StudentID, studentID)
	}
	if s.Filename != filename {
		t.Fatalf("filename = %q, want %q", s.Filename, filename)
	}
	if s.SizeBytes != len(content) {
		t.Fatalf("sizeBytes = %d, want %d", s.SizeBytes, len(content))
	}
	sum := sha256.Sum256(content)
	if s.SHA256 != hex.EncodeToString(sum[:]) {
		t.Fatalf("sha256 = %q, want %q", s.SHA256, hex.EncodeToString(sum[:]))
	}
	if s.Status != "queued" {
		t.Fatalf("status field = %q, want queued", s.Status)
	}
	if _, err := time.Parse(time.RFC3339, s.CreatedAt.Format(time.RFC3339)); err != nil {
		t.Fatalf("createdAt %v is not RFC 3339: %v", s.CreatedAt, err)
	}
	if loc := rec.Header().Get("Location"); loc != "/submissions/"+s.ID {
		t.Fatalf("Location = %q, want /submissions/%s", loc, s.ID)
	}
	return s
}

func TestCreateSubmission201(t *testing.T) {
	f := newFixture(t)
	content := []byte("void mergeSort(int* a, int n) {}\n")
	rec := f.submit(t, reqOpts{token: "t-alice", filename: "merge.cpp", content: content})
	wantCreated(t, rec, "u-alice", "merge.cpp", content)
	if n := f.store.CountSubmissions(); n != 1 {
		t.Fatalf("stored submissions = %d, want 1", n)
	}
}

func TestIdempotentReplay200(t *testing.T) {
	f := newFixture(t)
	content := []byte("int main() { return 0; }\n")
	first := f.submit(t, reqOpts{token: "t-alice", key: "retry-me", content: content})
	s1 := wantCreated(t, first, "u-alice", "sort.cpp", content)

	second := f.submit(t, reqOpts{token: "t-alice", key: "retry-me", content: content})
	if second.Code != http.StatusOK {
		t.Fatalf("replay status = %d, want 200 (body %q)", second.Code, second.Body.String())
	}
	if v := second.Header().Get("Idempotency-Replayed"); v != "true" {
		t.Fatalf("Idempotency-Replayed = %q, want true", v)
	}
	if loc := second.Header().Get("Location"); loc != "/submissions/"+s1.ID {
		t.Fatalf("replay Location = %q, want /submissions/%s", loc, s1.ID)
	}
	s2 := decodeSubmission(t, second)
	if s2.ID != s1.ID {
		t.Fatalf("replay id = %q, want original %q", s2.ID, s1.ID)
	}
	if n := f.store.CountSubmissions(); n != 1 {
		t.Fatalf("stored submissions = %d, want 1 (retry must not duplicate)", n)
	}
}

func TestReplayWinsOverClosedWindow(t *testing.T) {
	// Contract precedence step 8: a retry of an accepted upload returns the
	// original result even after the window closes. A retry never creates.
	f := newFixture(t)
	content := []byte("int main() {}\n")
	first := f.submit(t, reqOpts{token: "t-alice", key: "late-retry", content: content})
	s1 := wantCreated(t, first, "u-alice", "sort.cpp", content)

	lab, _ := f.store.Lab(labID)
	f.store.AddLab(&Lab{ID: lab.ID, ClassID: lab.ClassID, OpensAt: lab.OpensAt,
		ClosesAt: time.Now().UTC().Add(-time.Minute)})

	second := f.submit(t, reqOpts{token: "t-alice", key: "late-retry", content: content})
	if second.Code != http.StatusOK {
		t.Fatalf("replay after close status = %d, want 200 (body %q)", second.Code, second.Body.String())
	}
	if decodeSubmission(t, second).ID != s1.ID {
		t.Fatal("replay returned a different submission")
	}
	if n := f.store.CountSubmissions(); n != 1 {
		t.Fatalf("stored submissions = %d, want 1", n)
	}
}

func TestIdempotencyKeyReused409(t *testing.T) {
	f := newFixture(t)
	first := f.submit(t, reqOpts{token: "t-alice", key: "same-key", content: []byte("a")})
	wantCreated(t, first, "u-alice", "sort.cpp", []byte("a"))

	second := f.submit(t, reqOpts{token: "t-alice", key: "same-key", content: []byte("b")})
	p := wantProblem(t, second, http.StatusConflict, "idempotency-key-reused")
	if p.Instance == "" {
		t.Fatal("problem.instance is empty")
	}
	if n := f.store.CountSubmissions(); n != 1 {
		t.Fatalf("stored submissions = %d, want 1", n)
	}
}

func TestIdempotencyKeyScopedPerStudent(t *testing.T) {
	f := newFixture(t)
	content := []byte("int main() {}\n")
	first := f.submit(t, reqOpts{token: "t-alice", key: "shared-key", content: content})
	wantCreated(t, first, "u-alice", "sort.cpp", content)

	second := f.submit(t, reqOpts{token: "t-sam", key: "shared-key", content: content})
	wantCreated(t, second, "u-sam", "sort.cpp", content)
	if n := f.store.CountSubmissions(); n != 2 {
		t.Fatalf("stored submissions = %d, want 2", n)
	}
}

func TestUnauthenticated401(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: ""})
	wantProblem(t, rec, http.StatusUnauthorized, "unauthenticated")

	rec = f.submit(t, reqOpts{token: "not-a-real-token"})
	wantProblem(t, rec, http.StatusUnauthorized, "unauthenticated")
}

func TestMissingIdempotencyKey400(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-alice", omitKey: true})
	wantProblem(t, rec, http.StatusBadRequest, "invalid-idempotency-key")

	// Header explicitly set to empty string is "missing or empty".
	body := multipartBody(t, "sort.cpp", []byte("int main() {}\n"))
	req := httptest.NewRequest(http.MethodPost, "/labs/"+labID+"/submissions", body)
	req.Header.Set("Authorization", "Bearer t-alice")
	req.Header.Set("Idempotency-Key", "")
	req.Header.Set("Content-Type", body.contentType)
	emptyKeyRec := httptest.NewRecorder()
	f.mux.ServeHTTP(emptyKeyRec, req)
	wantProblem(t, emptyKeyRec, http.StatusBadRequest, "invalid-idempotency-key")

	rec = f.submit(t, reqOpts{token: "t-alice", key: strings.Repeat("k", 129)})
	wantProblem(t, rec, http.StatusBadRequest, "invalid-idempotency-key")
}

func TestBadLabID400(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-alice", labID: "not-a-uuid"})
	p := wantProblem(t, rec, http.StatusBadRequest, "validation-failed")
	if len(p.Errors) != 1 || p.Errors[0].Field != "labId" {
		t.Fatalf("errors = %+v, want one field error on labId", p.Errors)
	}
}

func TestUnsupportedMediaType415(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-alice", rawBody: "{}", contentType: "application/json"})
	wantProblem(t, rec, http.StatusUnsupportedMediaType, "unsupported-media-type")
}

func TestFileTooLarge413(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-alice", content: bytes.Repeat([]byte("a"), maxFileBytes+1)})
	p := wantProblem(t, rec, http.StatusRequestEntityTooLarge, "payload-too-large")
	if len(p.Errors) != 1 || p.Errors[0].Field != "file" {
		t.Fatalf("errors = %+v, want one field error on file", p.Errors)
	}
}

func TestBodyTooLarge413(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-alice", content: bytes.Repeat([]byte("a"), maxBodyBytes)})
	wantProblem(t, rec, http.StatusRequestEntityTooLarge, "payload-too-large")
}

func TestMissingFilePart400(t *testing.T) {
	f := newFixture(t)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("comment", "no file here")
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/labs/"+labID+"/submissions", &buf)
	req.Header.Set("Authorization", "Bearer t-alice")
	req.Header.Set("Idempotency-Key", "key-1")
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	f.mux.ServeHTTP(rec, req)

	p := wantProblem(t, rec, http.StatusBadRequest, "validation-failed")
	if len(p.Errors) != 1 || p.Errors[0].Field != "file" {
		t.Fatalf("errors = %+v, want one field error on file", p.Errors)
	}
}

func TestUnsupportedFileType422(t *testing.T) {
	f := newFixture(t)
	for _, name := range []string{"payload.sh", "sort.cpp.txt", "Makefile"} {
		rec := f.submit(t, reqOpts{token: "t-alice", filename: name})
		wantProblem(t, rec, http.StatusUnprocessableEntity, "unsupported-file-type")
	}
	// Case-insensitive: these must be accepted.
	for _, name := range []string{"SORT.CPP", "lib.HPP"} {
		rec := f.submit(t, reqOpts{token: "t-alice", filename: name, key: name})
		wantCreated(t, rec, "u-alice", name, []byte("#include <algorithm>\n"))
	}
}

func TestLabNotFound404(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-alice", labID: otherLab})
	wantProblem(t, rec, http.StatusNotFound, "lab-not-found")
}

func TestNonMemberGets404Not403(t *testing.T) {
	// Ownership rule: a lab that exists but is not visible to the caller is
	// indistinguishable from a lab that does not exist.
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-bob"})
	p := wantProblem(t, rec, http.StatusNotFound, "lab-not-found")

	missing := f.submit(t, reqOpts{token: "t-alice", labID: otherLab})
	m := wantProblem(t, missing, http.StatusNotFound, "lab-not-found")
	if p.Detail != m.Detail || p.Title != m.Title {
		t.Fatalf("non-member body %+v differs from missing-lab body %+v", p, m)
	}
	if n := f.store.CountSubmissions(); n != 0 {
		t.Fatalf("stored submissions = %d, want 0", n)
	}
}

func TestRoleNotAllowed403(t *testing.T) {
	f := newFixture(t)
	for _, token := range []string{"t-tina", "t-adam"} {
		rec := f.submit(t, reqOpts{token: token})
		wantProblem(t, rec, http.StatusForbidden, "role-not-allowed")
	}
	if n := f.store.CountSubmissions(); n != 0 {
		t.Fatalf("stored submissions = %d, want 0", n)
	}
}

func TestWindowNotOpen409(t *testing.T) {
	f := newFixture(t)
	lab, _ := f.store.Lab(labID)
	f.store.AddLab(&Lab{ID: lab.ID, ClassID: lab.ClassID,
		OpensAt: time.Now().UTC().Add(time.Hour), ClosesAt: time.Now().UTC().Add(2 * time.Hour)})
	rec := f.submit(t, reqOpts{token: "t-alice"})
	p := wantProblem(t, rec, http.StatusConflict, "submission-window-closed")
	if !strings.Contains(p.Detail, "not opened") {
		t.Fatalf("detail = %q, want it to say the window has not opened", p.Detail)
	}
}

func TestWindowClosed409(t *testing.T) {
	f := newFixture(t)
	lab, _ := f.store.Lab(labID)
	f.store.AddLab(&Lab{ID: lab.ID, ClassID: lab.ClassID,
		OpensAt: time.Now().UTC().Add(-2 * time.Hour), ClosesAt: time.Now().UTC().Add(-time.Hour)})
	rec := f.submit(t, reqOpts{token: "t-alice"})
	p := wantProblem(t, rec, http.StatusConflict, "submission-window-closed")
	if !strings.Contains(p.Detail, "closed") {
		t.Fatalf("detail = %q, want it to say the window has closed", p.Detail)
	}
}

func TestErrorPrecedenceAuthBeforeEverything(t *testing.T) {
	// An unauthenticated request with every other rule broken gets 401.
	f := newFixture(t)
	rec := f.submit(t, reqOpts{
		token: "", labID: "bad", omitKey: true,
		rawBody: "{}", contentType: "application/json",
	})
	wantProblem(t, rec, http.StatusUnauthorized, "unauthenticated")
}

func TestErrorPrecedenceLabIDBeforeKey(t *testing.T) {
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-alice", labID: "bad", omitKey: true})
	wantProblem(t, rec, http.StatusBadRequest, "validation-failed")
}

func TestErrorPrecedenceShapeBeforeOwnership(t *testing.T) {
	// A non-member with a broken upload shape gets the shape error, not 404:
	// the contract evaluates request shape before business rules.
	f := newFixture(t)
	rec := f.submit(t, reqOpts{token: "t-bob", filename: "payload.sh"})
	wantProblem(t, rec, http.StatusUnprocessableEntity, "unsupported-file-type")
}
