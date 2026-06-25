package handlers

import (
	"encoding/json"
	"net/http"

	"officehours/internal/auth"
	"officehours/internal/models"
)

type createSessionReq struct {
	AdvisorKey string `json:"advisor_key"`
	Kind       string `json:"kind"`
}

// CreateSession creates a session for the user. POST /api/sessions
func (s *Server) CreateSession(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	var req createSessionReq
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.AdvisorKey == "" {
		writeErr(w, http.StatusBadRequest, "advisor_key required")
		return
	}
	kind := req.Kind
	if kind == "" {
		kind = models.SessionKindOfficeHours
	}
	if kind != models.SessionKindOfficeHours && kind != models.SessionKindLearn {
		writeErr(w, http.StatusBadRequest, "invalid kind")
		return
	}

	title := titleForSession(s, kind, req.AdvisorKey)

	var sess models.Session
	err := s.DB.Pool.QueryRow(r.Context(),
		`insert into sessions (user_id, kind, advisor_key, title)
		 values ($1,$2,$3,$4)
		 returning id, user_id, kind, advisor_key, title, status, outcomes, created_at, concluded_at`,
		uid, kind, req.AdvisorKey, title).
		Scan(&sess.ID, &sess.UserID, &sess.Kind, &sess.AdvisorKey, &sess.Title, &sess.Status, &sess.Outcomes, &sess.CreatedAt, &sess.ConcludedAt)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Log a parcours/Logbook event for the session creation.
	if payload, perr := json.Marshal(map[string]any{
		"session_id":  sess.ID,
		"kind":        sess.Kind,
		"advisor_key": sess.AdvisorKey,
		"title":       sess.Title,
	}); perr == nil {
		_, _ = s.DB.Pool.Exec(r.Context(),
			`insert into events (user_id, kind, payload) values ($1,'session',$2)`,
			uid, payload)
	}

	// Advisor opens the session: enqueue an opening advisor/learn job (no user
	// message). The worker renders a short first assistant message.
	jobType := models.JobTypeAdvisor
	if kind == models.SessionKindLearn {
		jobType = models.JobTypeLearn
	}
	sid := sess.ID
	jobID, err := s.enqueueJob(r.Context(), jobType, uid, &sid, map[string]any{
		"opening": true,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sessionWithJob{Session: sess, JobID: jobID})
}

// sessionWithJob is the session create response: the session fields plus the
// id of the opening advisor job.
type sessionWithJob struct {
	models.Session
	JobID string `json:"job_id"`
}

// titleForSession derives a human title from the advisor/concept name.
func titleForSession(s *Server, kind, key string) string {
	dir := "advisors"
	if kind == models.SessionKindLearn {
		dir = "learn"
	}
	if def, err := loadDef(s, dir, key); err == nil && def.Name != "" {
		return def.Name
	}
	return key
}

// ListSessions returns the user's sessions (optionally filtered by kind).
// GET /api/sessions[?kind=]
func (s *Server) ListSessions(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	kind := r.URL.Query().Get("kind")

	sql := `select id, user_id, kind, advisor_key, title, status, outcomes, created_at, concluded_at
	        from sessions where user_id=$1`
	args := []any{uid}
	if kind != "" {
		sql += ` and kind=$2`
		args = append(args, kind)
	}
	sql += ` order by created_at desc`

	rows, err := s.DB.Pool.Query(r.Context(), sql, args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	out := []models.Session{}
	for rows.Next() {
		var sess models.Session
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.Kind, &sess.AdvisorKey, &sess.Title, &sess.Status, &sess.Outcomes, &sess.CreatedAt, &sess.ConcludedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = append(out, sess)
	}
	writeJSON(w, http.StatusOK, out)
}

// GetSession returns {session, messages, action_items}. GET /api/sessions/{id}
func (s *Server) GetSession(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id := r.PathValue("id")
	ctx := r.Context()

	var sess models.Session
	err := s.DB.Pool.QueryRow(ctx,
		`select id, user_id, kind, advisor_key, title, status, outcomes, created_at, concluded_at
		 from sessions where id=$1 and user_id=$2`, id, uid).
		Scan(&sess.ID, &sess.UserID, &sess.Kind, &sess.AdvisorKey, &sess.Title, &sess.Status, &sess.Outcomes, &sess.CreatedAt, &sess.ConcludedAt)
	if err != nil {
		writeErr(w, http.StatusNotFound, "session not found")
		return
	}

	msgRows, err := s.DB.Pool.Query(ctx,
		`select id, session_id, role, content, created_at from messages
		 where session_id=$1 order by created_at`, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer msgRows.Close()
	messages := []models.Message{}
	for msgRows.Next() {
		var m models.Message
		if err := msgRows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		messages = append(messages, m)
	}

	aiRows, err := s.DB.Pool.Query(ctx,
		`select id, user_id, session_id, title, horizon, rationale, program_ref, status, created_at
		 from action_items where session_id=$1 order by created_at`, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer aiRows.Close()
	items := []models.ActionItem{}
	for aiRows.Next() {
		var a models.ActionItem
		if err := aiRows.Scan(&a.ID, &a.UserID, &a.SessionID, &a.Title, &a.Horizon, &a.Rationale, &a.ProgramRef, &a.Status, &a.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, a)
	}

	// Latest queued|running job for this session (if any) -> pending_job.
	var pendingJob any // nil unless found
	var pjID, pjStatus string
	perr := s.DB.Pool.QueryRow(ctx,
		`select id, status from agent_jobs
		 where session_id=$1 and status in ('queued','running')
		 order by created_at desc limit 1`, id).Scan(&pjID, &pjStatus)
	if perr == nil {
		pendingJob = map[string]any{"id": pjID, "status": pjStatus}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session":      sess,
		"messages":     messages,
		"action_items": items,
		"pending_job":  pendingJob,
	})
}

type postMessageReq struct {
	Content string `json:"content"`
}

// PostMessage appends a user message and enqueues an advisor/learn job.
// POST /api/sessions/{id}/messages -> {job_id}
func (s *Server) PostMessage(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id := r.PathValue("id")
	var req postMessageReq
	if err := decode(r, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Content == "" {
		writeErr(w, http.StatusBadRequest, "content required")
		return
	}
	ctx := r.Context()

	// Verify ownership + get kind.
	var kind, status string
	err := s.DB.Pool.QueryRow(ctx,
		`select kind, status from sessions where id=$1 and user_id=$2`, id, uid).Scan(&kind, &status)
	if err != nil {
		writeErr(w, http.StatusNotFound, "session not found")
		return
	}
	if status == models.SessionStatusConcluded {
		writeErr(w, http.StatusConflict, "session is concluded")
		return
	}

	// Append the user message.
	_, err = s.DB.Pool.Exec(ctx,
		`insert into messages (session_id, role, content) values ($1,'user',$2)`, id, req.Content)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	jobType := models.JobTypeAdvisor
	if kind == models.SessionKindLearn {
		jobType = models.JobTypeLearn
	}
	sid := id
	jobID, err := s.enqueueJob(ctx, jobType, uid, &sid, map[string]any{
		"content": req.Content,
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"job_id": jobID})
}

// ConcludeSession enqueues a scorer job. POST /api/sessions/{id}/conclude -> {job_id}
func (s *Server) ConcludeSession(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	id := r.PathValue("id")
	ctx := r.Context()

	var status string
	err := s.DB.Pool.QueryRow(ctx,
		`select status from sessions where id=$1 and user_id=$2`, id, uid).Scan(&status)
	if err != nil {
		writeErr(w, http.StatusNotFound, "session not found")
		return
	}

	sid := id
	jobID, err := s.enqueueJob(ctx, models.JobTypeScorer, uid, &sid, map[string]any{})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"job_id": jobID})
}
