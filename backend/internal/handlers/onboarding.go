package handlers

import (
	"net/http"

	"officehours/internal/auth"
	"officehours/internal/models"
)

type onboardingReq struct {
	Text string `json:"text"`
}

// Onboarding creates/updates the profile company_text, ensures the default goal,
// and enqueues a diagnoser job. POST /api/onboarding -> {job_id}
func (s *Server) Onboarding(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var req onboardingReq
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Text == "" {
		writeErr(w, http.StatusBadRequest, "text required")
		return
	}
	ctx := r.Context()

	// Upsert the profile company_text and stamp onboarded_at on first submission.
	ct, err := s.DB.Pool.Exec(ctx,
		`update profiles
		 set company_text=$2, updated_at=now(),
		     onboarded_at=coalesce(onboarded_at, now())
		 where user_id=$1`, uid, req.Text)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ct.RowsAffected() == 0 {
		_, err = s.DB.Pool.Exec(ctx,
			`insert into profiles (user_id, company_text, onboarded_at) values ($1,$2,now())
			 on conflict (user_id) do update
			   set company_text=excluded.company_text, updated_at=now(),
			       onboarded_at=coalesce(profiles.onboarded_at, now())`,
			uid, req.Text)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Default goal "Start an Office Hours session" (created at onboarding).
	var exists bool
	_ = s.DB.Pool.QueryRow(ctx,
		`select exists(select 1 from goals where user_id=$1 and title=$2)`,
		uid, "Start an Office Hours session").Scan(&exists)
	if !exists {
		_, _ = s.DB.Pool.Exec(ctx,
			`insert into goals (user_id, title) values ($1,$2)`,
			uid, "Start an Office Hours session")
	}

	jobID, err := s.enqueueJob(ctx, models.JobTypeDiagnoser, uid, nil, map[string]any{
		"text": req.Text,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"job_id": jobID})
}
