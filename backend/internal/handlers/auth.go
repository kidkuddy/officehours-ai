package handlers

import (
	"net/http"

	"officehours/internal/auth"
	"officehours/internal/models"
)

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates email/password and returns {token, user}.
// POST /api/auth/login
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeErr(w, http.StatusBadRequest, "email and password required")
		return
	}
	if s.DB == nil {
		writeErr(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}

	var u models.User
	err := s.DB.Pool.QueryRow(r.Context(),
		`select id, email, password_hash, name, created_at from users where email=$1`,
		req.Email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !auth.CheckPassword(u.PasswordHash, req.Password) {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.IssueToken(s.Cfg.JWTSecret, u.ID, u.Email)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not issue token")
		return
	}

	// Also set httpOnly cookie for browser convenience.
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.TokenTTL.Seconds()),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  u,
	})
}

// Me returns the authenticated user + profile.
// GET /api/me
func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var u models.User
	err := s.DB.Pool.QueryRow(r.Context(),
		`select id, email, password_hash, name, created_at from users where id=$1`, uid).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, "user not found")
		return
	}
	prof := s.loadProfile(r, uid)
	writeJSON(w, http.StatusOK, map[string]any{"user": u, "profile": prof})
}
