package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// RestartApp runs: kubectl rollout restart deployment/app-<name> -n user-apps
func (h *Handler) RestartApp(w http.ResponseWriter, r *http.Request) {
	if !internalKeyOK(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	name := chi.URLParam(r, "name")
	out, err := runKubectl(r.Context(), "rollout", "restart",
		"deployment/app-"+name, "-n", "user-apps")
	if err != nil {
		log.Printf("ops: restart app=%s: %v: %s", name, err, out)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "restarted", "output": strings.TrimSpace(out)})
}

// ScaleApp runs: kubectl scale deployment/app-<name> --replicas=N -n user-apps
// Body: {"replicas": N} or {"replicas_delta": N} (adds to current count).
func (h *Handler) ScaleApp(w http.ResponseWriter, r *http.Request) {
	if !internalKeyOK(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	name := chi.URLParam(r, "name")

	var body struct {
		Replicas      *int `json:"replicas"`
		ReplicasDelta *int `json:"replicas_delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	var target int
	switch {
	case body.Replicas != nil:
		target = *body.Replicas
	case body.ReplicasDelta != nil:
		current, err := getCurrentReplicas(r.Context(), name)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "get replicas: " + err.Error()})
			return
		}
		target = current + *body.ReplicasDelta
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "replicas or replicas_delta required"})
		return
	}
	if target < 1 {
		target = 1
	}

	out, err := runKubectl(r.Context(), "scale",
		"deployment/app-"+name,
		fmt.Sprintf("--replicas=%d", target),
		"-n", "user-apps")
	if err != nil {
		log.Printf("ops: scale app=%s replicas=%d: %v: %s", name, target, err, out)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "scaled",
		"replicas": strconv.Itoa(target),
		"output":   strings.TrimSpace(out),
	})
}

// RollbackApp runs: kubectl rollout undo deployment/app-<name> -n user-apps
func (h *Handler) RollbackApp(w http.ResponseWriter, r *http.Request) {
	if !internalKeyOK(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	name := chi.URLParam(r, "name")
	out, err := runKubectl(r.Context(), "rollout", "undo",
		"deployment/app-"+name, "-n", "user-apps")
	if err != nil {
		log.Printf("ops: rollback app=%s: %v: %s", name, err, out)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "rolled-back", "output": strings.TrimSpace(out)})
}

// getCurrentReplicas reads .spec.replicas from the live deployment.
func getCurrentReplicas(ctx context.Context, appName string) (int, error) {
	out, err := runKubectl(ctx,
		"get", "deployment", "app-"+appName,
		"-n", "user-apps",
		"-o", "jsonpath={.spec.replicas}",
	)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 1, nil // safe default if jsonpath returns empty
	}
	return n, nil
}

func runKubectl(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// internalKeyOK checks X-Internal-Key header. Passes through in dev (no key set).
func internalKeyOK(r *http.Request) bool {
	key := os.Getenv("INTERNAL_API_KEY")
	if key == "" {
		return true
	}
	return r.Header.Get("X-Internal-Key") == key
}
