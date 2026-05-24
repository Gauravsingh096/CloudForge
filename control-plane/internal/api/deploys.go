package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Deploy struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Image     string    `json:"image"`
	CommitSHA string    `json:"commit_sha"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) ListDeploys(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, project_id, image, commit_sha, status, created_at
		 FROM deploys WHERE project_id = $1 ORDER BY created_at DESC LIMIT 20`,
		projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	deploys := []Deploy{}
	for rows.Next() {
		var d Deploy
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Image, &d.CommitSHA, &d.Status, &d.CreatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		deploys = append(deploys, d)
	}
	writeJSON(w, http.StatusOK, deploys)
}

func (h *Handler) CreateDeploy(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		Image     string `json:"image"`
		CommitSHA string `json:"commit_sha"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if input.ProjectID == "" || input.Image == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project_id and image are required"})
		return
	}

	var d Deploy
	err := h.db.QueryRowContext(r.Context(),
		`INSERT INTO deploys (project_id, image, commit_sha, status)
		 VALUES ($1, $2, $3, 'pending')
		 RETURNING id, project_id, image, commit_sha, status, created_at`,
		input.ProjectID, input.Image, input.CommitSHA,
	).Scan(&d.ID, &d.ProjectID, &d.Image, &d.CommitSHA, &d.Status, &d.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

func (h *Handler) GetDeploy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var d Deploy
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, project_id, image, commit_sha, status, created_at FROM deploys WHERE id = $1`, id,
	).Scan(&d.ID, &d.ProjectID, &d.Image, &d.CommitSHA, &d.Status, &d.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "deploy not found"})
		return
	}
	writeJSON(w, http.StatusOK, d)
}
