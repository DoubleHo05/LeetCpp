package api

import (
	"encoding/json"
	"net/http"
)

// Problem is an RFC 9457 problem details object.
type Problem struct {
	Type     string       `json:"type"`
	Title    string       `json:"title"`
	Status   int          `json:"status"`
	Detail   string       `json:"detail,omitempty"`
	Instance string       `json:"instance,omitempty"`
	Errors   []FieldError `json:"errors,omitempty"`
}

// FieldError names one offending request field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// problemBase is the type URI namespace from docs/contract/; it is an
// identifier namespace, not a website.
const problemBase = "https://leetc.example/problems/"

func writeProblem(w http.ResponseWriter, r *http.Request, status int, slug, title, detail string, fieldErrors ...FieldError) {
	p := Problem{
		Type:     problemBase + slug,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: r.URL.Path,
		Errors:   fieldErrors,
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}
