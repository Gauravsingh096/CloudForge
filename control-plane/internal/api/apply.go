package api

import (
	"log"
	"net/http"

	"github.com/Gauravsingh096/cloudforge/control-plane/internal/deploys"
	"github.com/go-chi/chi/v5"
)

// ApplyDeploy is called by GitHub Actions after the image is pushed to ghcr.io.
// It triggers kubectl apply for the deploy and streams the result.
func (h *Handler) ApplyDeploy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	deployer := deploys.New(h.db)
	if err := deployer.Apply(r.Context(), id); err != nil {
		log.Printf("apply deploy %s failed: %v", id, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"deploy_id": id,
		"status":    "running",
	})
}
