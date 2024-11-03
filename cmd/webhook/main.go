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
	"github.com/samilton/gitsaver/pkgs/github"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger
var token string

func initLogger() {
	// Create a production logger configuration
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	sugar = logger.Sugar()
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-GitHub-Delivery")
	logger := logger.With(zap.String("request_id", requestID))

	// Verify HTTP method
	if r.Method != http.MethodPost {
		logger.Warn("invalid HTTP method",
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the webhook secret from environment variable
	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		logger.Error("webhook secret not configured")
		http.Error(w, "Webhook secret not configured", http.StatusInternalServerError)
		return
	}

	// Read the request body
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("failed to read request body",
			zap.Error(err),
			zap.String("remote_addr", r.RemoteAddr),
		)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Verify signature
	signature := r.Header.Get("X-Hub-Signature-256")
	if !verifySignature(payload, signature, webhookSecret) {
		logger.Warn("invalid webhook signature",
			zap.String("remote_addr", r.RemoteAddr),
		)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse the event type
	eventType := r.Header.Get("X-GitHub-Event")
	logger = logger.With(zap.String("event_type", eventType))
	if eventType != "push" {
		// Acknowledge but ignore non-push events
		logger.Info("ignoring non-push event")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse the payload
	var pushEvent github.WebhookPayload
	if err := json.Unmarshal(payload, &pushEvent); err != nil {
		sugar.Error("failed to parse JSON body",
			zap.Error(err),
			zap.String("payload", string(payload)),
		)
		http.Error(w, "Failed to parse webhook payload", http.StatusBadRequest)
		return
	}

	logger = logger.With(
		zap.String("repository", pushEvent.Repository.FullName),
		zap.String("ref", pushEvent.Ref),
	)
	// Check if this is a push to main/master branch
	if !isMainBranch(pushEvent.Ref) {
		logger.Info("ignoring push to non-main branch")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Create backup
	if err := backupRepository(pushEvent, logger); err != nil {
		logger.Error("failed to backup repository",
			zap.Error(err),
		)
		http.Error(w, "Failed to backup repository", http.StatusInternalServerError)
		return
	}

	logger.Info("webhook processed successfully")
	w.WriteHeader(http.StatusOK)
}

func isMainBranch(ref string) bool {
	return ref == "refs/heads/main" || ref == "refs/heads/master"
}

func backupRepository(event github.WebhookPayload, logger *zap.Logger) error {
	// Validate repository owner and name
	if strings.Contains(event.Repository.Owner.Name, "/") || strings.Contains(event.Repository.Owner.Name, "\\") || strings.Contains(event.Repository.Owner.Name, "..") {
		return fmt.Errorf("invalid repository owner name: %s", event.Repository.Owner.Name)
	}
	if strings.Contains(event.Repository.Name, "/") || strings.Contains(event.Repository.Name, "\\") || strings.Contains(event.Repository.Name, "..") {
		return fmt.Errorf("invalid repository name: %s", event.Repository.Name)
	}

	// Create backup directory with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join("backups", event.Repository.Owner.Name,
		event.Repository.Name, timestamp)

	logger = logger.With(zap.String("backup_dir", backupDir))

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	logger.Info("created backup directory")

	// Clone the repository
	logger.Info("starting repository clone")

	_, err := git.PlainClone(backupDir, false, &git.CloneOptions{
		URL: event.Repository.CloneURL,
		Auth: &githttp.BasicAuth{
			Username: "x-access-token",
			Password: token,
		},
		Progress: os.Stdout,
	})

	if err != nil {
		logger.Error("clone failed",
			zap.Error(err),
		)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	logger.Info("repository backup completed successfully")

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
	appID := int64(1044603)
	installationID := int64(56702972)
	privateKeyPath := "private-key.pem"

	initLogger()
	defer logger.Sync()

	logger.Info("starting application")
	app, err := github.NewGitHubApp(appID, installationID, privateKeyPath)
	if err != nil {
		logger.Fatal("failed to create GitHub app",
			zap.Error(err),
		)
	}

	token, err = app.GetInstallationToken()
	if err != nil {
		logger.Fatal("failed to get installation token",
			zap.Error(err),
		)
	}
	logger.Info("Got installation token", zap.String("token", token))

	if err := os.MkdirAll("backups", 0755); err != nil {
		logger.Fatal("failed to create backups directory",
			zap.Error(err),
		)
		return
	}

	http.HandleFunc("/", handleWebhook)
	addr := ":8080"
	logger.Info("starting server",
		zap.String("address", addr),
	)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("server failed to start",
			zap.Error(err),
		)
	}
}
