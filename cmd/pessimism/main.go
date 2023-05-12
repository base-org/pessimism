package main

import (
	"context"
	"os"
	"sync"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/engine"
	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/config"

	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/logging"
)

const (
	cfgPath = "config.env"
)

func initializeAndRunServer(ctx context.Context, cfgPath config.FilePath,
	etlMan *pipeline.Manager, engineMan *engine.Manager) (*server.Server, func(), error) {
	cfg := config.NewConfig(cfgPath)

	apiService := service.New(ctx, cfg.SvcConfig, etlMan, engineMan)
	handler, err := handlers.New(ctx, apiService)
	if err != nil {
		return nil, nil, err
	}
	server, cleanup, err := server.New(ctx, cfg.ServerConfig, handler)
	if err != nil {
		return nil, nil, err
	}
	return server, func() {
		cleanup()
	}, nil
}

func main() {
	appWg := &sync.WaitGroup{}

	appCtx, appCtxCancel := context.WithCancel(context.Background())

	cfg := config.NewConfig(cfgPath)

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logger := logging.WithContext(appCtx)
	logger.Info("Bootstrapping pessimsim monitoring application")

	engineManager, shutDownEngine := engine.NewManager()
	etlManager, shutDownETL := pipeline.NewManager(appCtx, engineManager.Transit())

	logger.Info("Starting and running risk engine manager instance")
	engineCtx, engineCtxCancel := context.WithCancel(appCtx)

	appWg.Add(1)
	go func() {
		defer appWg.Done()

		etlManager.EventLoop(appCtx)
	}()

	logger.Info("Starting and running ETL manager instance")

	appWg.Add(1)
	go func() {
		defer appWg.Done()

		if err := engineManager.EventLoop(engineCtx); err != nil {
			logger.Error("engine manager event loop crashed", zap.Error(err))
		}
	}()

	appWg.Add(1)
	go func() {
		defer appWg.Done()

		server, shutDownServer, err := initializeAndRunServer(appCtx, cfgPath, etlManager, engineManager)

		if err != nil {
			logger.Error("Error obtained trying to start server", zap.Error(err))
			panic(err)
		}

		server.Stop(func() {
			logger.Info("Shutting down pessimism application")

			engineCtxCancel() // Shutdown risk engine event-loop
			shutDownEngine()  // Shutdown risk engine subsystem
			logger.Info("Shutdown risk engine subsystem")

			appCtxCancel() // Shutdown ETL subsystem event loops
			shutDownETL()  // Shutdown ETL subsystem
			logger.Info("Shutdown ETL subsystem")

			shutDownServer() // Shutdown API server
			logger.Info("Shutdown API server")

			logger.Info("Shutdown context")
		})
	}()

	logger.Debug("Waiting for all application threads to end")
	appWg.Wait()
	logger.Info("Successful pessimism shutdown")
	os.Exit(0)
}
