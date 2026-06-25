// Package handlers implements the HTTP API in BUILD_SPEC §4. Routes are wired
// onto a *http.ServeMux under the /api base path by Register.
package handlers

import (
	"encoding/json"
	"net/http"

	"officehours/internal/auth"
	"officehours/internal/config"
	"officehours/internal/db"
)

// Server holds shared dependencies for the handlers.
type Server struct {
	Cfg *config.Config
	DB  *db.DB
}

// New constructs a handler Server.
func New(cfg *config.Config, database *db.DB) *Server {
	return &Server{Cfg: cfg, DB: database}
}

// Register wires every API route onto the mux (already stripped of /api).
func Register(mux *http.ServeMux, cfg *config.Config, database *db.DB) {
	s := New(cfg, database)
	secret := cfg.JWTSecret

	// Public.
	mux.HandleFunc("POST /auth/login", s.Login)

	// Authenticated.
	mux.HandleFunc("GET /me", auth.Require(secret, s.Me))
	mux.HandleFunc("POST /onboarding", auth.Require(secret, s.Onboarding))
	mux.HandleFunc("GET /profile", auth.Require(secret, s.GetProfile))
	mux.HandleFunc("GET /signals", auth.Require(secret, s.GetSignals))
	mux.HandleFunc("GET /dashboard", auth.Require(secret, s.GetDashboard))
	mux.HandleFunc("GET /goals", auth.Require(secret, s.GetGoals))
	mux.HandleFunc("GET /events", auth.Require(secret, s.ListEvents))
	mux.HandleFunc("GET /advisors", auth.Require(secret, s.GetAdvisors))
	mux.HandleFunc("GET /learn", auth.Require(secret, s.GetLearn))

	mux.HandleFunc("POST /sessions", auth.Require(secret, s.CreateSession))
	mux.HandleFunc("GET /sessions", auth.Require(secret, s.ListSessions))
	mux.HandleFunc("GET /sessions/{id}", auth.Require(secret, s.GetSession))
	mux.HandleFunc("POST /sessions/{id}/messages", auth.Require(secret, s.PostMessage))
	mux.HandleFunc("POST /sessions/{id}/conclude", auth.Require(secret, s.ConcludeSession))

	mux.HandleFunc("POST /documents", auth.Require(secret, s.UploadDocument))
	mux.HandleFunc("GET /documents", auth.Require(secret, s.ListDocuments))

	mux.HandleFunc("GET /jobs/{id}", auth.Require(secret, s.GetJob))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// decode reads a JSON body into v.
func decode(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
