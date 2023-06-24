package main

import (
	"context"
	"os"

	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/config"
)

const (
	// cfgPath ... env file path
	cfgPath = "config.env"
)

// main ... Application driver
func main() {

	cfg := config.NewConfig(cfgPath) // Load env vars
	ctx := context.Background()      // Create context

	pessimism, kill, err := app.NewPessimismApp(ctx, cfg)
	logger := logging.WithContext(ctx)

	if err != nil {
		logger.Fatal("Error creating pessimism application", zap.Error(err))
	}

	logger.Info("Starting pessimism application")
	if err := pessimism.Start(); err != nil {
		logger.Fatal("Error starting pessimism application", zap.Error(err))
	}

	if cfg.IsBootstrap() {
		logger.Debug("Bootstrapping application state")

		sessions, err := fetchBootSessions(cfg.BootStrapPath)
		if err != nil {
			logger.Fatal("Error loading bootstrap file", zap.Error(err))
		}

		if err := pessimism.BootStrap(sessions); err != nil {
			logger.Fatal("Error bootstrapping application state", zap.Error(err))
		}

		logger.Debug("Application state successfully bootstrapped")
	}

	pessimism.ListenForShutdown(func() {
		// TODO - Add shutdown for subsystem manager
		if err != nil {
			logger.Error("Error shutting down subsystems", zap.Error(err))
		}

		kill()
	})

	logger.Debug("Waiting for all application threads to end")

	logger.Info("Successful pessimism shutdown")
	os.Exit(0)
}
