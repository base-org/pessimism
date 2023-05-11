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
	wg := &sync.WaitGroup{}

	appCtx, ctxCancel := context.WithCancel(context.Background())

	cfg := config.NewConfig(cfgPath)

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logger := logging.WithContext(appCtx)

	logger.Info("bootstrapping pessimsim monitoring application")

	engineManager, shutDownEngine := engine.NewManager()

	logger.Info("starting and running ETL manager instance")

	etlManager, shutDownETL := pipeline.NewManager(appCtx, engineManager.Transit(), wg)

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := engineManager.EventLoop(appCtx); err != nil {
			logger.Error("engine manager event loop crashed", zap.Error(err))
		}
	}()

	logger.Info("starting and running risk engine manager instance")

	wg.Add(1)
	go func() {
		defer wg.Done()

		etlManager.EventLoop(appCtx)
	}()

	go func() {
		server, shutDownServer, err := initializeAndRunServer(appCtx, cfgPath, etlManager, engineManager)

		if err != nil {
			logger.Error("Error obtained trying to start server", zap.Error(err))
			panic(err)
		}

		server.Stop(func() {
			shutDownETL()
			shutDownEngine()
			shutDownServer()
			ctxCancel()
		})
	}()

	wg.Wait()
	logger.Info("Ending pessimism application run")
	os.Exit(0)
}
