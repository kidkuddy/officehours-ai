package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"officehours/internal/agent"
	"officehours/internal/auth"
	"officehours/internal/models"
)

// loadDef is a small helper to load an agent/concept definition by key.
func loadDef(s *Server, dir, key string) (*agent.Definition, error) {
	return agent.LoadByKey(filepath.Join(s.Cfg.SeedDir, dir), key)
}

// userCollection is the per-user Data Room collection name.
func userCollection(uid string) string {
	return "user-" + uid
}

// UploadDocument stores an uploaded file and enqueues a rag index job.
// POST /api/documents (multipart: file) -> {document}
func (s *Server) UploadDocument(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	ctx := r.Context()

	if !s.Cfg.Features.DataRoom.Enabled {
		writeErr(w, http.StatusForbidden, "data room disabled")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "file field required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !accepted(s.Cfg.Features.DataRoom.Accept, ext) {
		writeErr(w, http.StatusUnsupportedMediaType,
			fmt.Sprintf("file type %s not accepted", ext))
		return
	}

	collection := userCollection(uid)
	dir := filepath.Join(s.Cfg.UploadDir, uid)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not create upload dir")
		return
	}
	// Unique on-disk name.
	stored := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
	dst := filepath.Join(dir, stored)
	out, err := os.Create(dst)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not store file")
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		writeErr(w, http.StatusInternalServerError, "could not write file")
		return
	}
	out.Close()

	mime := header.Header.Get("Content-Type")

	var doc models.Document
	err = s.DB.Pool.QueryRow(ctx,
		`insert into documents (user_id, collection, filename, mime, path)
		 values ($1,$2,$3,$4,$5)
		 returning id, user_id, collection, filename, mime, path, created_at`,
		uid, collection, header.Filename, mime, dst).
		Scan(&doc.ID, &doc.UserID, &doc.Collection, &doc.Filename, &doc.Mime, &doc.Path, &doc.CreatedAt)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Enqueue rag index job for this document.
	_, err = s.enqueueJob(ctx, "rag_index", uid, nil, map[string]any{
		"document_id": doc.ID,
		"path":        dst,
		"collection":  collection,
	})
	if err != nil {
		// Non-fatal: doc is stored; log via error field is enough.
		fmt.Fprintf(os.Stderr, "documents: enqueue index failed: %v\n", err)
	}

	writeJSON(w, http.StatusOK, map[string]any{"document": doc})
}

// ListDocuments returns the user's documents. GET /api/documents
func (s *Server) ListDocuments(w http.ResponseWriter, r *http.Request) {
	uid := auth.UserID(r.Context())
	rows, err := s.DB.Pool.Query(r.Context(),
		`select id, user_id, collection, filename, mime, path, created_at
		 from documents where user_id=$1 order by created_at desc`, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	out := []models.Document{}
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.UserID, &d.Collection, &d.Filename, &d.Mime, &d.Path, &d.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		out = append(out, d)
	}
	writeJSON(w, http.StatusOK, out)
}

func accepted(list []string, ext string) bool {
	if len(list) == 0 {
		return true
	}
	for _, a := range list {
		if strings.EqualFold(a, ext) {
			return true
		}
	}
	return false
}
