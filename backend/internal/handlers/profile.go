package handlers

import (
	"encoding/json"
	"net/http"

	"officehours/internal/auth"
	"officehours/internal/models"
)

// profileJSON is the API shape for a profile (stage_evidence as raw JSON).
type profileJSON struct {
	UserID        string          `json:"user_id"`
	CompanyText   string          `json:"company_text"`
	Stage         string          `json:"stage"`
	StageEvidence json.RawMessage `json:"stage_evidence"`
	CreatedAt     any             `json:"created_at"`
	UpdatedAt     any             `json:"updated_at"`
	Onboarded     bool            `json:"onboarded"`
}

// loadProfile fetches the profile for a user, or nil if absent.
func (s *Server) loadProfile(r *http.Request, uid string) *profileJSON {
	var p models.Profile
	var evidence []byte
	err := s.DB.Pool.QueryRow(r.Context(),
		`select user_id, company_text, stage, stage_evidence, created_at, updated_at, onboarded_at
		 from profiles where user_id=$1`, uid).
		Scan(&p.UserID, &p.CompanyText, &p.Stage, &evidence, &p.CreatedAt, &p.UpdatedAt, &p.OnboardedAt)
	if err != nil {
		return nil
	}
	if len(evidence) == 0 {
		evidence = []byte("[]")
	}
	return &profileJSON{
		UserID: p.UserID, CompanyText: p.CompanyText, Stage: p.Stage,
		StageEvidence: json.RawMessage(evidence),
		CreatedAt:     p.CreatedAt, UpdatedAt: p.UpdatedAt,
		Onboarded:     p.OnboardedAt != nil,
	}
}

// loadSignals returns the user's signals with subscores as raw JSON.
func (s *Server) loadSignals(r *http.Request, uid string) ([]map[string]any, error) {
	rows, err := s.DB.Pool.Query(r.Context(),
		`select id, user_id, name, score, subscores, rationale, floor_triggered, updated_at
		 from signals where user_id=$1 order by name`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var sig models.Signal
		var sub []byte
		if err := rows.Scan(&sig.ID, &sig.UserID, &sig.Name, &sig.Score, &sub, &sig.Rationale, &sig.FloorTriggered, &sig.UpdatedAt); err != nil {
			return nil, err
		}
		if len(sub) == 0 {
			sub = []byte("[]")
		}
		out = append(out, map[string]any{
			"id": sig.ID, "user_id": sig.UserID, "name": sig.Name, "score": sig.Score,
			"subscores": json.RawMessage(sub), "rationale": sig.Rationale,
			"floor_triggered": sig.FloorTriggered, "updated_at": sig.UpdatedAt,
		})
	}
	return out, rows.Err()
}

// GetProfile returns profile + signals preview. GET /api/profile
func (s *Server) GetProfile(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	prof := s.loadProfile(r, uid)
	if prof == nil {
		writeErr(w, http.StatusNotFound, "profile not found")
		return
	}
	sigs, err := s.loadSignals(r, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": prof, "signals": sigs})
}

// GetSignals returns the user's signals. GET /api/signals
func (s *Server) GetSignals(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	sigs, err := s.loadSignals(r, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sigs)
}

// GetGoals returns the user's goals. GET /api/goals
func (s *Server) GetGoals(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := s.DB.Pool.Query(r.Context(),
		`select id, user_id, title, description, status, source_session_id, created_at, done_at
		 from goals where user_id=$1 order by created_at`, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	out := []models.Goal{}
	for rows.Next() {
		var g models.Goal
		if err := rows.Scan(&g.ID, &g.UserID, &g.Title, &g.Description, &g.Status, &g.SourceSessionID, &g.CreatedAt, &g.DoneAt); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = append(out, g)
	}
	writeJSON(w, http.StatusOK, out)
}

// GetDashboard returns {stage, signals, stats, parcours}. GET /api/dashboard
func (s *Server) GetDashboard(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	ctx := r.Context()

	prof := s.loadProfile(r, uid)
	stage := models.StageIdeation
	if prof != nil {
		stage = prof.Stage
	}

	sigs, err := s.loadSignals(r, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Stats.
	var sessionCount, goalsOpen, goalsDone, actionItems, docCount int
	_ = s.DB.Pool.QueryRow(ctx, `select count(*) from sessions where user_id=$1`, uid).Scan(&sessionCount)
	_ = s.DB.Pool.QueryRow(ctx, `select count(*) from goals where user_id=$1 and status='open'`, uid).Scan(&goalsOpen)
	_ = s.DB.Pool.QueryRow(ctx, `select count(*) from goals where user_id=$1 and status='done'`, uid).Scan(&goalsDone)
	_ = s.DB.Pool.QueryRow(ctx, `select count(*) from action_items where user_id=$1`, uid).Scan(&actionItems)
	_ = s.DB.Pool.QueryRow(ctx, `select count(*) from documents where user_id=$1`, uid).Scan(&docCount)

	// Parcours timeline (events, newest first).
	rows, err := s.DB.Pool.Query(ctx,
		`select id, user_id, kind, payload, created_at from events
		 where user_id=$1 order by created_at desc limit 100`, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	parcours := []map[string]any{}
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.Kind, &e.Payload, &e.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		payload := e.Payload
		if len(payload) == 0 {
			payload = []byte("{}")
		}
		parcours = append(parcours, map[string]any{
			"id": e.ID, "kind": e.Kind, "payload": json.RawMessage(payload),
			"created_at": e.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stage":   stage,
		"signals": sigs,
		"stats": map[string]int{
			"sessions": sessionCount, "goals_open": goalsOpen,
			"goals_done": goalsDone, "action_items": actionItems, "documents": docCount,
		},
		"parcours": parcours,
	})
}
