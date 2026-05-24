package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestBuildImageName(t *testing.T) {
	tests := []struct {
		fullName string
		sha      string
		want     string
	}{
		{"Gauravsingh096/cloudforge", "abc1234567", "ghcr.io/gauravsingh096/cloudforge:abc1234"},
		{"owner/my-app", "deadbeef12", "ghcr.io/owner/my-app:deadbeef"},
		{"UPPERCASE/REPO", "1234567890", "ghcr.io/uppercase/repo:1234567"},
		{"org/sample-app", "f3887f640c", "ghcr.io/org/sample-app:f3887f6"},
	}
	for _, tc := range tests {
		got := buildImageName(tc.fullName, tc.sha)
		if got != tc.want {
			t.Errorf("buildImageName(%q, %q) = %q, want %q", tc.fullName, tc.sha, got, tc.want)
		}
	}
}

func TestVerifyGithubSignature_EmptySecret(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", "")
	if !verifyGithubSignature([]byte("any-payload"), "any-sig") {
		t.Error("expected true when WEBHOOK_SECRET is empty (auth disabled)")
	}
}

func TestVerifyGithubSignature_Valid(t *testing.T) {
	secret := "cloudforge-test-secret"
	payload := []byte(`{"ref":"refs/heads/main","after":"abc123"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	t.Setenv("WEBHOOK_SECRET", secret)
	if !verifyGithubSignature(payload, sig) {
		t.Error("expected valid HMAC signature to pass")
	}
}

func TestVerifyGithubSignature_Invalid(t *testing.T) {
	t.Setenv("WEBHOOK_SECRET", "cloudforge-test-secret")
	if verifyGithubSignature([]byte("payload"), "sha256=badhash") {
		t.Error("expected tampered signature to fail")
	}
}

func TestVerifyGithubSignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"ref":"refs/heads/main"}`)
	mac := hmac.New(sha256.New, []byte("correct-secret"))
	mac.Write(payload)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	t.Setenv("WEBHOOK_SECRET", "wrong-secret")
	if verifyGithubSignature(payload, sig) {
		t.Error("expected signature computed with wrong secret to fail")
	}
}
