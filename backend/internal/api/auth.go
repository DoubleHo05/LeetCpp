package api

import "net/http"

// Role of a caller as known to the platform.
type Role string

const (
	RoleStudent Role = "student"
	RoleTA      Role = "ta"
	RoleAdmin   Role = "admin"
)

// User is the authenticated caller.
type User struct {
	ID       string
	Role     Role
	ClassIDs []string
}

// MemberOf reports whether the user belongs to the given class.
func (u *User) MemberOf(classID string) bool {
	for _, c := range u.ClassIDs {
		if c == classID {
			return true
		}
	}
	return false
}

// Authenticator resolves the caller of a request. ok is false when the
// request carries no credentials the server recognizes.
type Authenticator interface {
	Authenticate(r *http.Request) (*User, bool)
}

// BearerAuthenticator resolves "Authorization: Bearer <token>" against a
// static table. It is a stand-in for real sessions until the auth milestone
// lands; tests inject their own table.
type BearerAuthenticator struct {
	Tokens map[string]*User
}

const bearerPrefix = "Bearer "

// Authenticate implements Authenticator.
func (a BearerAuthenticator) Authenticate(r *http.Request) (*User, bool) {
	h := r.Header.Get("Authorization")
	if len(h) <= len(bearerPrefix) || h[:len(bearerPrefix)] != bearerPrefix {
		return nil, false
	}
	u, ok := a.Tokens[h[len(bearerPrefix):]]
	return u, ok
}
