package api

import (
	"sync"
	"time"
)

// Lab is the assignment a submission belongs to. Submissions are accepted
// while now is inside [OpensAt, ClosesAt].
type Lab struct {
	ID       string
	ClassID  string
	OpensAt  time.Time
	ClosesAt time.Time
}

// Submission is the representation returned by POST /labs/{labId}/submissions.
// Field names match docs/contract/post-labs-labid-submissions.md.
type Submission struct {
	ID        string    `json:"id"`
	LabID     string    `json:"labId"`
	StudentID string    `json:"studentId"`
	Filename  string    `json:"filename"`
	SizeBytes int       `json:"sizeBytes"`
	SHA256    string    `json:"sha256"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store is the persistence the submit endpoint needs.
type Store interface {
	Lab(id string) (*Lab, bool)
	// FindIdempotent looks up a previous request under
	// (studentID, labID, key). ok is false when the key was never used.
	FindIdempotent(studentID, labID, key string) (fingerprint string, sub *Submission, ok bool)
	Create(sub *Submission, studentID, labID, key, fingerprint string)
}

type idemRecord struct {
	fingerprint string
	sub         *Submission
}

// InMemoryStore is a Store safe for concurrent use. The Postgres-backed
// store replaces it in a later milestone; contract tests use this one.
type InMemoryStore struct {
	mu   sync.Mutex
	labs map[string]*Lab
	subs map[string]*Submission
	idem map[string]idemRecord
}

// NewInMemoryStore returns an empty store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		labs: map[string]*Lab{},
		subs: map[string]*Submission{},
		idem: map[string]idemRecord{},
	}
}

// AddLab inserts or replaces a lab.
func (s *InMemoryStore) AddLab(l *Lab) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.labs[l.ID] = l
}

// Lab implements Store.
func (s *InMemoryStore) Lab(id string) (*Lab, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.labs[id]
	return l, ok
}

// FindIdempotent implements Store.
func (s *InMemoryStore) FindIdempotent(studentID, labID, key string) (string, *Submission, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.idem[s.idemKey(studentID, labID, key)]
	return rec.fingerprint, rec.sub, ok
}

// Create implements Store.
func (s *InMemoryStore) Create(sub *Submission, studentID, labID, key, fingerprint string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[sub.ID] = sub
	s.idem[s.idemKey(studentID, labID, key)] = idemRecord{fingerprint: fingerprint, sub: sub}
}

// CountSubmissions reports how many submissions exist (used by tests).
func (s *InMemoryStore) CountSubmissions() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

func (s *InMemoryStore) idemKey(studentID, labID, key string) string {
	return studentID + "\x00" + labID + "\x00" + key
}
