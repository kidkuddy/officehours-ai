package handlers

import (
	"encoding/json"
	"net/http"

	"officehours/internal/auth"
	"officehours/internal/models"
)

// ListEvents returns the current user's events (Logbook / Mon Parcours),
// newest first. GET /api/events -> [{id, kind, payload, created_at}]
func (s *Server) ListEvents(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := s.DB.Pool.Query(r.Context(),
		`select id, kind, payload, created_at from events
		 where user_id=$1 order by created_at desc`, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	out := []map[string]any{}
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.Kind, &e.Payload, &e.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		payload := e.Payload
		if len(payload) == 0 {
			payload = []byte("{}")
		}
		out = append(out, map[string]any{
			"id":         e.ID,
			"kind":       e.Kind,
			"payload":    json.RawMessage(payload),
			"created_at": e.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
