package main

import (
	"context"
	"os"

	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
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
	logger := logging.WithContext(ctx)

	l1Client, err := client.NewEthClient(ctx, cfg.L1RpcEndpoint)
	if err != nil {
		logging.WithContext(ctx).Fatal("Error creating L1 client", zap.Error(err))
	}

	l2Client, err := client.NewEthClient(ctx, cfg.L2RpcEndpoint)
	if err != nil {
		logging.WithContext(ctx).Fatal("Error creating L1 client", zap.Error(err))
	}

	ss := state.NewMemState()

	ctx = app.InitializeContext(ctx, ss, l1Client, l2Client)

	pessimism, shutDown, err := app.NewPessimismApp(ctx, cfg)

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

	pessimism.ListenForShutdown(shutDown)

	logger.Debug("Waiting for all application threads to end")

	logger.Info("Successful pessimism shutdown")
	os.Exit(0)
}
