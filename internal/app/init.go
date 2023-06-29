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
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/state"
	"github.com/base-org/pessimism/internal/subsystem"
	"go.uber.org/zap"
)

// InitializeContext ... Performs dependency injection to build context struct
func InitializeContext(ctx context.Context, ss state.Store,
	l1Client, l2Client client.EthClientInterface) context.Context {
	ctx = context.WithValue(
		ctx, core.State, ss)

	ctx = context.WithValue(
		ctx, core.L1Client, l1Client)

	return context.WithValue(
		ctx, core.L2Client, l2Client)
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
func InitializeServer(ctx context.Context, cfg *config.Config, m subsystem.Manager) (*server.Server, func(), error) {
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

// InitializeAlerting ... Performs dependency injection to build alerting struct
func InitializeAlerting(ctx context.Context, cfg *config.Config) alert.Manager {
	sc := client.NewSlackClient(cfg.SlackURL)
	return alert.NewManager(ctx, sc)
}

// InitalizeETL ... Performs dependency injection to build etl struct
func InitalizeETL(ctx context.Context, transit chan core.InvariantInput) pipeline.Manager {
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

	return engine.NewManager(ctx, re, am, store, transit)
}

// NewPessimismApp ... Performs dependency injection to build app struct
func NewPessimismApp(ctx context.Context, cfg *config.Config) (*Application, func(), error) {
	mSvr, mShutDown, err := InitializeMetrics(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	alrt := InitializeAlerting(ctx, cfg)
	engine := InitializeEngine(ctx, alrt.Transit())
	etl := InitalizeETL(ctx, engine.Transit())

	m := subsystem.NewManager(ctx, etl, engine, alrt)

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
