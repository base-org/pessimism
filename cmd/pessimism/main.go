package main

import (
	"context"
	"os"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/state"
	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/subsystem"

	"github.com/base-org/pessimism/internal/config"

	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/logging"
)

const (
	// cfgPath ... env file path
	cfgPath = "config.env"
)

// initializeServer ... Performs dependency injection to build server struct
func initializeServer(ctx context.Context, cfg *config.Config,
	m subsystem.Manager) (*server.Server, func(), error) {
	apiService := service.New(ctx, cfg.SvcConfig, m)
	handler, err := handlers.New(ctx, apiService)
	if err != nil {
		return nil, nil, err
	}

	server, cleanup, err := server.New(ctx, cfg.ServerConfig, handler)
	if err != nil {
		return nil, nil, err
	}
	return server, cleanup, nil
}

/*
	Subsystem initialization functions
*/

// initializeAlerting ... Performs dependency injection to build alerting struct
func initializeAlerting(ctx context.Context, cfg *config.Config) alert.AlertingManager {
	sc := client.NewSlackClient(cfg.SlackURL)
	return alert.NewManager(ctx, sc)
}

// initalizeETL ... Performs dependency injection to build etl struct
func initalizeETL(ctx context.Context, transit chan core.InvariantInput) pipeline.Manager {
	compRegistry := registry.NewRegistry()
	analyzer := pipeline.NewAnalyzer(compRegistry)
	store := pipeline.NewEtlStore()
	dag := pipeline.NewComponentGraph()

	return pipeline.NewManager(ctx, analyzer, compRegistry, store, dag, transit)
}

// initializeEngine ... Performs dependency injection to build engine struct
func initializeEngine(ctx context.Context, transit chan core.Alert) engine.Manager {
	store := engine.NewSessionStore()
	am := engine.NewAddressingMap()
	re := engine.NewHardCodedEngine()

	return engine.NewManager(ctx, re, am, store, transit)
}

// main ... Application driver
func main() {
	ctx := context.WithValue(
		context.Background(), state.Default, state.NewMemState())

	cfg := config.NewConfig(cfgPath) // Load env vars

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logger := logging.WithContext(ctx)
	logger.Info("Bootstrapping pessimsim monitoring application")

	alrt := initializeAlerting(ctx, cfg)
	eng := initializeEngine(ctx, alrt.Transit())
	etl := initalizeETL(ctx, eng.Transit())

	m := subsystem.NewManager(ctx, etl, eng, alrt)
	srver, shutdownServer, err := initializeServer(ctx, cfg, m)
	if err != nil {
		logger.Error("Error initializing server", zap.Error(err))
		os.Exit(1)
	}

	pess := app.New(ctx, m, srver)

	logger.Info("Starting pessimism application")
	if err := pess.Start(); err != nil {
		logger.Error("Error starting pessimism application", zap.Error(err))
		os.Exit(1)
	}

	pess.ListenForShutdown(func() {
		err := m.Shutdown()
		if err != nil {
			logger.Error("Error shutting down subsystems", zap.Error(err))
		}

		shutdownServer()
	})

	logger.Debug("Waiting for all application threads to end")

	logger.Info("Successful pessimism shutdown")
	os.Exit(0)
}
