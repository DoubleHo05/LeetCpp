package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DoubleHo05/LeetCpp/backend/internal/api"
)

func main() {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	store := api.NewInMemoryStore()
	auth := api.BearerAuthenticator{Tokens: map[string]*api.User{
		// Placeholder dev tokens until real session auth lands in the auth
		// milestone. Never ship these; they exist so `go run ./cmd/api` has
		// a caller to exercise.
		"dev-alice": {ID: "u-alice", Role: api.RoleStudent, ClassIDs: []string{"class-a"}},
	}}

	mux := http.NewServeMux()
	mux.Handle("POST /labs/{labId}/submissions", api.NewSubmitHandler(store, auth, time.Now))

	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
