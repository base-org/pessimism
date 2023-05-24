package main

import (
	"context"
	"os"
	"sync"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/state"
	"go.uber.org/zap"

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
func initializeServer(ctx context.Context, cfg *config.Config, alertMan alert.AlertingManager,
	etlMan pipeline.Manager, engineMan engine.Manager) (*server.Server, func(), error) {
	apiService := service.New(ctx, cfg.SvcConfig, alertMan, etlMan, engineMan)
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

// initializeAlerting ... Performs dependency injection to build alerting struct
func initializeAlerting(ctx context.Context, cfg *config.Config) (alert.AlertingManager, func()) {
	sc := client.NewSlackClient(cfg.SlackURL)
	return alert.NewManager(ctx, sc)
}

// main ... Application driver
func main() {
	appWg := &sync.WaitGroup{}
	appCtx, appCtxCancel := context.WithCancel(context.Background())

	appCtx = context.WithValue(appCtx, state.Default, state.NewMemState())

	cfg := config.NewConfig(cfgPath) // Load env vars

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logger := logging.WithContext(appCtx)
	logger.Info("Bootstrapping pessimsim monitoring application")
	compRegistry := registry.NewRegistry()

	alertingManager, shutdownAlerting := initializeAlerting(appCtx, cfg)

	engineManager, shutDownEngine := engine.NewManager(appCtx, alertingManager.Transit())
	etlManager, shutDownETL := pipeline.NewManager(appCtx, compRegistry, engineManager.Transit())

	logger.Info("Starting and running risk engine manager instance")
	engineCtx, engineCtxCancel := context.WithCancel(appCtx)

	appWg.Add(1)
	go func() { // EtlManager driver thread
		defer appWg.Done()

		if err := etlManager.EventLoop(appCtx); err != nil {
			logger.Error("etl manager event loop error", zap.Error(err))
		}
	}()

	logger.Info("Starting and running ETL manager instance")

	appWg.Add(1)
	go func() { // EngineManager driver thread
		defer appWg.Done()

		if err := engineManager.EventLoop(engineCtx); err != nil {
			logger.Error("engine manager event loop error", zap.Error(err))
		}
	}()

	appWg.Add(1)
	go func() { // AlertManager driver thread
		defer appWg.Done()

		if err := alertingManager.EventLoop(engineCtx); err != nil {
			logger.Error("alert manager event loop error", zap.Error(err))
		}
	}()

	appWg.Add(1)
	go func() { // ApiServer driver thread
		defer appWg.Done()

		apiServer, shutDownServer, err := initializeServer(appCtx, cfg, alertingManager,
			etlManager, engineManager)

		if err != nil {
			logger.Error("Error obtained trying to start server", zap.Error(err))
			panic(err)
		}

		apiServer.Stop(func() {
			logger.Info("Shutting down pessimism application")

			shutdownAlerting() // Shutdown alerting subsystem

			engineCtxCancel() // Shutdown risk engine event-loop
			shutDownEngine()  // Shutdown risk engine subsystem
			logger.Info("Shutdown risk engine subsystem")

			appCtxCancel() // Shutdown ETL subsystem event loops
			shutDownETL()  // Shutdown ETL subsystem
			logger.Info("Shutdown ETL subsystem")

			shutDownServer() // Shutdown API server
			logger.Info("Shutdown API server")
		})
	}()

	logger.Debug("Waiting for all application threads to end")
	appWg.Wait()
	logger.Info("Successful pessimism shutdown")
	os.Exit(0)
}
