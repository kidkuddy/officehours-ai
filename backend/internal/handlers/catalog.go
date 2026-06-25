package handlers

import (
	"net/http"
	"path/filepath"

	"officehours/internal/agent"
)

type catalogItem struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetAdvisors lists enabled advisors from /seed/advisors/*.md. GET /api/advisors
func (s *Server) GetAdvisors(w http.ResponseWriter, r *http.Request) {
	defs, err := agent.LoadDir(filepath.Join(s.Cfg.SeedDir, "advisors"))
	if err != nil {
		writeJSON(w, http.StatusOK, []catalogItem{})
		return
	}
	out := []catalogItem{}
	for _, d := range defs {
		if !d.Enabled {
			continue
		}
		out = append(out, catalogItem{Key: d.Key, Name: d.Name, Description: d.Description})
	}
	writeJSON(w, http.StatusOK, out)
}

// GetLearn lists enabled learn concepts from /seed/learn/*.md, gated by
// features.yaml learn.enabled. GET /api/learn
func (s *Server) GetLearn(w http.ResponseWriter, r *http.Request) {
	if !s.Cfg.Features.Learn.Enabled {
		writeJSON(w, http.StatusOK, []catalogItem{})
		return
	}
	defs, err := agent.LoadDir(filepath.Join(s.Cfg.SeedDir, "learn"))
	if err != nil {
		writeJSON(w, http.StatusOK, []catalogItem{})
		return
	}
	out := []catalogItem{}
	for _, d := range defs {
		if !d.Enabled {
			continue
		}
		out = append(out, catalogItem{Key: d.Key, Name: d.Name, Description: d.Description})
	}
	writeJSON(w, http.StatusOK, out)
}
