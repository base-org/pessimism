package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/subsystem"
	"go.uber.org/zap"
)

// Application ... Pessimism app struct
type Application struct {
	ctx context.Context

	sub    subsystem.Manager
	server *server.Server
}

// New ... Initializer
func New(ctx context.Context, sub subsystem.Manager, server *server.Server) *Application {
	return &Application{
		ctx:    ctx,
		sub:    sub,
		server: server,
	}
}

// Start ... Starts the application
func (a *Application) Start() error {
	// Spawn subsystem event loop routines
	a.sub.StartEventRoutines(a.ctx)

	// Start the API server
	a.server.Start()
	return nil
}

// ListenForShutdown ... Handles and listens or shutdown
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
