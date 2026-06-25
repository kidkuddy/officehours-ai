package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"officehours/internal/auth"
	"officehours/internal/models"
)

// enqueueJob inserts a queued agent_job and returns its id.
func (s *Server) enqueueJob(ctx context.Context, jobType, userID string, sessionID *string, input map[string]any) (string, error) {
	in, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	var id string
	err = s.DB.Pool.QueryRow(ctx,
		`insert into agent_jobs (type, user_id, session_id, status, input)
		 values ($1,$2,$3,'queued',$4) returning id`,
		jobType, userID, sessionID, in).Scan(&id)
	return id, err
}

// GetJob returns {status, output, error}. GET /api/jobs/{id}
func (s *Server) GetJob(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id := r.PathValue("id")

	var j models.AgentJob
	err := s.DB.Pool.QueryRow(r.Context(),
		`select id, type, user_id, session_id, status, input, output, error, created_at, started_at, finished_at
		 from agent_jobs where id=$1 and user_id=$2`, id, uid).
		Scan(&j.ID, &j.Type, &j.UserID, &j.SessionID, &j.Status, &j.Input, &j.Output, &j.Error, &j.CreatedAt, &j.StartedAt, &j.FinishedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, "job not found")
		return
	}
	output := j.Output
	if len(output) == 0 {
		output = []byte("{}")
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":     j.ID,
		"status": j.Status,
		"output": json.RawMessage(output),
		"error":  j.Error,
	})
}
