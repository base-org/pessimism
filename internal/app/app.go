package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/subsystem"
	"go.uber.org/zap"
)

// BootSession ... Application wrapper for InvRequestParams
type BootSession = models.InvRequestParams

// Application ... Pessimism app struct
type Application struct {
	cfg     *config.Config
	ctx     context.Context
	metrics metrics.Metricer

	sub    subsystem.Manager
	server *server.Server
}

// New ... Initializer
func New(ctx context.Context, cfg *config.Config,
	sub subsystem.Manager, server *server.Server, stats metrics.Metricer) *Application {
	return &Application{
		ctx:     ctx,
		cfg:     cfg,
		sub:     sub,
		server:  server,
		metrics: stats,
	}
}

// Start ... Starts the application
func (a *Application) Start() error {
	// Start metrics server
	a.metrics.Start()

	// Spawn subsystem event loop routines
	a.sub.StartEventRoutines(a.ctx)

	// Start the API server
	a.server.Start()

	metrics.WithContext(a.ctx).RecordUp()

	return nil
}

// ListenForShutdown ... Handles and listens for shutdown
func (a *Application) ListenForShutdown(stop func()) {
	done := <-a.End() // Blocks until an OS signal is received

	logging.WithContext(a.ctx).
		Info("Received shutdown OS signal", zap.String("signal", done.String()))
	stop()
}

// End ... Returns a channel that will receive an OS signal
func (a *Application) End() <-chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}

// BootStrap ... Bootstraps the application
func (a *Application) BootStrap(sessions []BootSession) error {
	logger := logging.WithContext(a.ctx)

	for _, session := range sessions {
		inv, err := registry.GetInvariant(a.ctx, session.InvariantType(), session.SessionParams)
		if err != nil {
			return err
		}

		pollInterval, err := a.cfg.SvcConfig.GetPollIntervalForNetwork(session.NetworkType())
		if err != nil {
			return err
		}

		pConfig := session.GeneratePipelineConfig(pollInterval, inv.InputType())
		sConfig := session.SessionConfig()

		sUUID, err := a.sub.StartInvSession(pConfig, sConfig)
		if err != nil {
			return err
		}

		logger.Info("invariant session started",
			zap.String(core.SUUIDKey, sUUID.String()))
	}
	return nil
}
