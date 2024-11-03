package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the webhook secret from environment variable
	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		http.Error(w, "Webhook secret not configured", http.StatusInternalServerError)
		return
	}

	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Verify signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !verifySignature(payload, signature, webhookSecret) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse the event type
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "push" {
		// Acknowledge but ignore non-push events
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the payload
	var pushEvent PushEvent
	if err := json.Unmarshal(payload, &pushEvent); err != nil {
		http.Error(w, "Failed to parse webhook payload", http.StatusBadRequest)
		return
	}

	// Check if this is a push to main/master branch
	if !isMainBranch(pushEvent.Ref) {
		fmt.Printf("Ignoring push to non-main branch: %s\n", pushEvent.Ref)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Create backup
	if err := backupRepository(pushEvent); err != nil {
		fmt.Printf("Failed to backup repository: %v\n", err)
		http.Error(w, "Failed to backup repository", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func isMainBranch(ref string) bool {
	return ref == "refs/heads/main" || ref == "refs/heads/master"
}

func backupRepository(event PushEvent) error {
	// Create backup directory with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join("backups", event.Repository.Owner.Name,
		event.Repository.Name, timestamp)

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get GitHub token for authentication
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Clone the repository
	fmt.Printf("Cloning repository to %s\n", backupDir)

	_, err := git.PlainClone(backupDir, false, &git.CloneOptions{
		URL: event.Repository.CloneURL,
		Auth: &githttp.BasicAuth{
			Username: "git", // This can be anything except empty
			Password: token,
		},
		Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	fmt.Printf("Successfully backed up repository to %s\n", backupDir)
	return nil
}

func verifySignature(payload []byte, signature string, secret string) bool {
	if len(signature) == 0 || !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	signature = strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

func main() {
	appID := int64(YOUR_APP_ID)
	if err := os.MkdirAll("backups", 0755); err != nil {
		fmt.Printf("Failed to create backups directory: %v\n", err)
		return
	}

	http.HandleFunc("/webhook", handleWebhook)
	fmt.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
