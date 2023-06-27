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

	"github.com/base-org/pessimism/internal/metrics"
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
	ethClient := client.NewEthClient()
	apiService := service.New(ctx, cfg.SvcConfig, m, ethClient)
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
func initializeAlerting(ctx context.Context, cfg *config.Config) alert.Manager {
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

func initializeContext(cfg *config.Config) {
	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())
	metrics.NewStatsd(cfg.MetricsConfig)
}

// main ... Application driver
func main() {
	ctx := context.WithValue(
		context.Background(), state.Default, state.NewMemState())

	cfg := config.NewConfig(cfgPath) // Load env vars

	// Initialize logger and metrics
	initializeContext(cfg)
	stats := metrics.WithContext(ctx)
	stats.Incr("mainRuntime.start", []string{}, 1)

	logger := logging.WithContext(ctx)
	logger.Info("Bootstrapping pessimism monitoring application")

	alrt := initializeAlerting(ctx, cfg)
	eng := initializeEngine(ctx, alrt.Transit())
	etl := initalizeETL(ctx, eng.Transit())

	m := subsystem.NewManager(ctx, etl, eng, alrt)
	srver, shutdownServer, err := initializeServer(ctx, cfg, m)
	if err != nil {
		logger.Error("Error initializing server", zap.Error(err))
		os.Exit(1)
	}

	pessimism := app.New(ctx, cfg, m, srver)

	logger.Info("Starting pessimism application")
	if err := pessimism.Start(); err != nil {
		logger.Error("Error starting pessimism application", zap.Error(err))
		os.Exit(1)
	}

	if cfg.IsBootstrap() {
		logger.Debug("Bootstrapping application state")

		sessions, err := fetchBootSessions(cfg.BootStrapPath)
		if err != nil {
			logger.Error("Error loading bootstrap file", zap.Error(err))
			panic(err)
		}

		if err := pessimism.BootStrap(sessions); err != nil {
			logger.Error("Error bootstrapping application state", zap.Error(err))
			panic(err)
		}

		logger.Debug("Application state successfully bootstrapped")
	}

	pessimism.ListenForShutdown(func() {
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
