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
func InitalizeETL(ctx context.Context, transit chan core.InvariantInput, metr metrics.Metricer) pipeline.Manager {
	compRegistry := registry.NewRegistry()
	analyzer := pipeline.NewAnalyzer(compRegistry)
	store := pipeline.NewEtlStore()
	dag := pipeline.NewComponentGraph()

	return pipeline.NewManager(ctx, analyzer, compRegistry, store, dag, metr, transit)
}

// InitializeEngine ... Performs dependency injection to build engine struct
func InitializeEngine(ctx context.Context, transit chan core.Alert, metr metrics.Metricer) engine.Manager {
	store := engine.NewSessionStore()
	am := engine.NewAddressingMap()
	re := engine.NewHardCodedEngine()

	return engine.NewManager(ctx, re, am, store, metr, transit)
}

// NewPessimismApp ... Performs dependency injection to build app struct
func NewPessimismApp(ctx context.Context, cfg *config.Config, metr metrics.Metricer) (*Application, func(), error) {
	if metr == nil {
		metr = metrics.NoopMetrics
	}

	alrt := InitializeAlerting(ctx, cfg)
	engine := InitializeEngine(ctx, alrt.Transit(), metr)
	etl := InitalizeETL(ctx, engine.Transit(), metr)

	m := subsystem.NewManager(ctx, etl, engine, alrt)

	svr, shutDown, err := InitializeServer(ctx, cfg, m)
	if err != nil {
		return nil, nil, err
	}

	appShutDown := func() {
		shutDown()

		if err := m.Shutdown(); err != nil {
			logging.WithContext(ctx).Error("error shutting down subsystems", zap.Error(err))
		}
	}

	return New(ctx, cfg, m, svr, metr), appShutDown, nil
}
