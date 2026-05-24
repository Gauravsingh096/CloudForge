package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
)

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

	log.Printf("github webhook: push event received (%d bytes)", len(body))
	// TODO(week3): parse push payload, extract repo + sha, trigger deploy pipeline
	w.WriteHeader(http.StatusAccepted)
}

func verifyGithubSignature(body []byte, signature string) bool {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		return true // skip verification in dev if secret not set
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
