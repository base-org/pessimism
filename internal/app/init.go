package app

import (
	"context"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	e_registry "github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/state"
	"github.com/base-org/pessimism/internal/subsystem"

	"go.uber.org/zap"
)

// InitializeContext ... Performs dependency injection to build context struct
func InitializeContext(ctx context.Context, ss state.Store, cb *client.Bundle) context.Context {
	ctx = context.WithValue(
		ctx, core.State, ss)

	return context.WithValue(
		ctx, core.Clients, cb)
}

// InitializeMetrics ... Performs dependency injection to build metrics struct
func InitializeMetrics(ctx context.Context, cfg *config.Config) (metrics.Metricer, func(), error) {
	if !cfg.MetricsConfig.Enabled {
		return metrics.NoopMetrics, func() {}, nil
	}

	server, cleanup, err := metrics.New(ctx, cfg.MetricsConfig)
	if err != nil {
		return nil, nil, err
	}

	return server, cleanup, nil
}

// InitializeServer ... Performs dependency injection to build server struct
func InitializeServer(ctx context.Context, cfg *config.Config, m *subsystem.Manager) (*server.Server, func(), error) {
	apiService := service.New(ctx, m)
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

// InitializeAlerting ... Performs dependency injection to build alerting struct
func InitializeAlerting(ctx context.Context, cfg *config.Config) (alert.Manager, error) {
	if err := cfg.IngestAlertConfig(); err != nil {
		return nil, err
	}

	clientMap := alert.NewRoutingDirectory(cfg.AlertConfig)

	return alert.NewManager(ctx, cfg.AlertConfig, clientMap), nil
}

// InitializeETL ... Performs dependency injection to build etl struct
func InitializeETL(ctx context.Context, transit chan core.HeuristicInput) pipeline.Manager {
	compRegistry := registry.NewRegistry()
	analyzer := pipeline.NewAnalyzer(compRegistry)
	store := pipeline.NewEtlStore()
	dag := pipeline.NewComponentGraph()

	return pipeline.NewManager(ctx, analyzer, compRegistry, store, dag, transit)
}

// InitializeEngine ... Performs dependency injection to build engine struct
func InitializeEngine(ctx context.Context, transit chan core.Alert) engine.Manager {
	store := engine.NewSessionStore()
	am := engine.NewAddressingMap()
	re := engine.NewHardCodedEngine()
	it := e_registry.NewHeuristicTable()

	return engine.NewManager(ctx, re, am, store, it, transit)
}

// NewPessimismApp ... Performs dependency injection to build app struct
func NewPessimismApp(ctx context.Context, cfg *config.Config) (*Application, func(), error) {
	mSvr, mShutDown, err := InitializeMetrics(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	alerting, err := InitializeAlerting(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	engine := InitializeEngine(ctx, alerting.Transit())
	etl := InitializeETL(ctx, engine.Transit())

	m := subsystem.NewManager(ctx, cfg.SystemConfig, etl, engine, alerting)

	svr, shutDown, err := InitializeServer(ctx, cfg, m)
	if err != nil {
		return nil, nil, err
	}

	appShutDown := func() {
		shutDown()
		mShutDown()
		if err := m.Shutdown(); err != nil {
			logging.WithContext(ctx).Error("error shutting down subsystems", zap.Error(err))
		}
	}

	return New(ctx, cfg, m, svr, mSvr), appShutDown, nil
}
