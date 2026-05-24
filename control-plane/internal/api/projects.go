package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
	Subdomain string    `json:"subdomain"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, name, repo_url, subdomain, created_at FROM projects ORDER BY created_at DESC`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	projects := []Project{}
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.RepoURL, &p.Subdomain, &p.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		projects = append(projects, p)
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		RepoURL string `json:"repo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if input.Name == "" || input.RepoURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and repo_url are required"})
		return
	}

	var p Project
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO projects (name, repo_url, subdomain)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, repo_url, subdomain, created_at`,
		input.Name, input.RepoURL, input.Name,
	).Scan(&p.ID, &p.Name, &p.RepoURL, &p.Subdomain, &p.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p Project
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, name, repo_url, subdomain, created_at FROM projects WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.RepoURL, &p.Subdomain, &p.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}
	writeJSON(w, http.StatusOK, p)
}
