package main

import (
	"context"
	"os"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/base-org/pessimism/cmd/doc"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/state"
)

const (
	// cfgPath ... env file path
	cfgPath = "config.env"
)

// main ... Application driver
func main() {
	ctx := context.Background() // Create context
	logger := logging.WithContext(ctx)

	app := cli.NewApp()
	app.Name = "pessimism"
	app.Usage = "Pessimism Application"
	app.Description = "A monitoring service that allows for " +
		"Op-Stack and EVM compatible blockchains to be continuously assessed for real-time threats"
	app.Action = RunPessimism
	app.Commands = []cli.Command{
		{
			Name:        "doc",
			Subcommands: doc.Subcommands,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal("Error running application", zap.Error(err))
	}
}

func RunPessimism(_ *cli.Context) error {
	cfg := config.NewConfig(cfgPath) // Load env vars
	ctx := context.Background()

	// Init logger
	logging.NewLogger(cfg.LoggerConfig, string(cfg.Environment))
	logger := logging.WithContext(ctx)

	// Init stats server
	stats := initializeMetrics(ctx, cfg)

	l1Client, err := client.NewEthClient(ctx, cfg.L1RpcEndpoint)
	if err != nil {
		logger.Fatal("Error creating L1 client", zap.Error(err))
		return err
	}

	l2Client, err := client.NewEthClient(ctx, cfg.L2RpcEndpoint)
	if err != nil {
		logger.Fatal("Error creating L1 client", zap.Error(err))
		return err
	}

	ss := state.NewMemState()

	ctx = app.InitializeContext(ctx, ss, l1Client, l2Client)

	pessimism, shutDown, err := app.NewPessimismApp(ctx, cfg, stats)

	if err != nil {
		logger.Fatal("Error creating pessimism application", zap.Error(err))
		return err
	}

	logger.Info("Starting pessimism application")
	if err := pessimism.Start(); err != nil {
		logger.Fatal("Error starting pessimism application", zap.Error(err))
		return err
	}

	if cfg.IsBootstrap() {
		logger.Debug("Bootstrapping application state")

		sessions, err := fetchBootSessions(cfg.BootStrapPath)
		if err != nil {
			logger.Fatal("Error loading bootstrap file", zap.Error(err))
			return err
		}

		if err := pessimism.BootStrap(sessions); err != nil {
			logger.Fatal("Error bootstrapping application state", zap.Error(err))
			return err
		}

		logger.Debug("Application state successfully bootstrapped")
	}

	pessimism.ListenForShutdown(shutDown)

	logger.Debug("Waiting for all application threads to end")

	logger.Info("Successful pessimism shutdown")
	return nil
}

func initializeMetrics(ctx context.Context, cfg *config.Config) *metrics.Metrics {
	logger := logging.WithContext(ctx)

	met := metrics.NewMetrics()
	if !cfg.MetricsConfig.Enabled {
		logger.Info("Metrics server disabled")
		return nil
	}

	if cfg.MetricsConfig.Enabled {
		go func() {
			if err := met.Serve(ctx, cfg.MetricsConfig); err != nil {
				logger.Fatal("Error starting metrics server", zap.Error(err))
				panic(err)
			}
		}()

		logger.Info("Metrics server started",
			zap.String("host", cfg.MetricsConfig.Host), zap.Uint64("port", cfg.MetricsConfig.Port))
	}
	return met
}
