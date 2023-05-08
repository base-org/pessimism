package main

import (
	"context"
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
	handler, err := handlers.New(apiService)
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

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewConfig(cfgPath)

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logger := logging.WithContext(appCtx)

	logger.Info("pessimism boot up")

	engineManager := engine.NewManager()

	logger.Info("starting and running ETL manager")

	etlManager, shutDownETL := pipeline.NewManager(appCtx, engineManager.Transit())

	wg.Add(1)

	go func() {
		defer wg.Done()

		engineManager.EventLoop(appCtx)

	}()

	logger.Info("starting and running risk engine manager")

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
			shutDownServer()
		})
	}()

	wg.Wait()
}
