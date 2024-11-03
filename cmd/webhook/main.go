package main

import (
	"fmt"
	"os"

	"github.com/samilton/gitsaver/internal/config"
	"github.com/samilton/gitsaver/internal/server"
	"github.com/samilton/gitsaver/pkgs/github"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
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
	//sugar := logger.Sugar()
}

func main() {
	initLogger()
	defer logger.Sync()

	logger.Info("starting application")

	cfg, err := config.Load(logger)
	if err != nil {
		logger.Fatal("failed to load config",
			zap.Error(err),
		)
	}

	app, err := github.NewGitHubApp(cfg.GitHub.AppID, cfg.GitHub.InstallationID, cfg.GitHub.PrivateKeyPath)
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

	server := server.NewServer(cfg, logger, token)

	if err := server.Start(); err != nil {
		logger.Fatal("server failed to start",
			zap.Error(err),
		)
	}
}
