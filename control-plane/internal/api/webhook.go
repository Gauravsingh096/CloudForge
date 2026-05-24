package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type githubPushEvent struct {
	Ref        string `json:"ref"`
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
		HTMLURL  string `json:"html_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"head_commit"`
}

func (h *Handler) GithubWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}

	if !verifyGithubSignature(body, r.Header.Get("X-Hub-Signature-256")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	if event != "push" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var payload githubPushEvent
	if err := json.Unmarshal(body, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	// only deploy on push to main/master
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	if branch != "main" && branch != "master" {
		log.Printf("webhook: skipping push to branch %s", branch)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	repoURL := payload.Repository.HTMLURL
	sha := payload.After

	log.Printf("webhook: push to %s branch=%s sha=%s", repoURL, branch, sha[:7])

	// find matching project in DB
	var projectID string
	err = h.db.QueryRowContext(r.Context(),
		`SELECT id FROM projects WHERE repo_url = $1`, repoURL,
	).Scan(&projectID)
	if err != nil {
		log.Printf("webhook: no project found for repo %s", repoURL)
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "no project registered for this repo — create one via POST /api/projects first",
		})
		return
	}

	// create deploy record
	image := buildImageName(payload.Repository.FullName, sha)
	var deployID string
	err = h.db.QueryRowContext(r.Context(),
		`INSERT INTO deploys (project_id, image, commit_sha, status)
		 VALUES ($1, $2, $3, 'pending')
		 RETURNING id`,
		projectID, image, sha,
	).Scan(&deployID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	log.Printf("webhook: created deploy %s for project %s", deployID, projectID)
	writeJSON(w, http.StatusAccepted, map[string]string{
		"deploy_id":  deployID,
		"project_id": projectID,
		"image":      image,
		"status":     "pending",
	})
}

func buildImageName(fullName, sha string) string {
	// fullName = "Gauravsingh096/my-app" → "gauravsingh096/my-app"
	parts := strings.SplitN(strings.ToLower(fullName), "/", 2)
	if len(parts) != 2 {
		return fullName
	}
	return "ghcr.io/" + parts[0] + "/" + parts[1] + ":" + sha[:7]
}

func verifyGithubSignature(body []byte, signature string) bool {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
