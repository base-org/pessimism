package subsystem

import (
	"context"
	"sync"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Manager ... Subsystem manager interface
type Manager interface {
	StartEventRoutines(ctx context.Context)
	StartInvSession(cfg *core.PipelineConfig, invCfg *core.SessionConfig) (core.SUUID, error)
	Shutdown() error
}

// manager ... Subsystem manager struct
type manager struct {
	ctx context.Context

	etl  pipeline.Manager
	eng  engine.Manager
	alrt alert.Manager

	*sync.WaitGroup
}

// NewManager ... Initializer for the subsystem manager
func NewManager(ctx context.Context, etl pipeline.Manager, eng engine.Manager,
	alrt alert.Manager,
) Manager {
	return &manager{
		ctx:       ctx,
		etl:       etl,
		eng:       eng,
		alrt:      alrt,
		WaitGroup: &sync.WaitGroup{},
	}
}

// Shutdown ... Shuts down all subsystems in primary data flow order
// Ie. ETL -> Engine -> Alert
func (m *manager) Shutdown() error {
	if err := m.etl.Shutdown(); err != nil {
		return err
	}

	if err := m.eng.Shutdown(); err != nil {
		return err
	}

	return m.alrt.Shutdown()
}

// StartEventRoutines ... Starts the event loop routines for the subsystems
func (m *manager) StartEventRoutines(ctx context.Context) {
	logger := logging.WithContext(ctx)

	m.Add(1)
	go func() { // EngineManager driver thread
		defer m.Done()

		if err := m.eng.EventLoop(); err != nil {
			logger.Error("engine manager event loop error", zap.Error(err))
		}
	}()

	m.Add(1)
	go func() { // AlertManager driver thread
		defer m.Done()

		if err := m.alrt.EventLoop(); err != nil {
			logger.Error("alert manager event loop error", zap.Error(err))
		}
	}()

	m.Add(1)
	go func() { // ETL driver thread
		defer m.Done()

		if err := m.alrt.EventLoop(); err != nil {
			logger.Error("ETL manager event loop error", zap.Error(err))
		}
	}()
}

// StartInvSession ... Deploys an invariant session
func (m *manager) StartInvSession(cfg *core.PipelineConfig, invCfg *core.SessionConfig) (core.SUUID, error) {
	logger := logging.WithContext(m.ctx)

	// NOTE: This is a temporary solution
	// Parameterized preloading should be defined in an invariant register definition
	if invCfg.Type == core.WithdrawalEnforcement {
		invCfg.Params["args"] = []interface{}{"WithdrawalProven(bytes32,address,address)"}
		invCfg.Params["address"] = invCfg.Params["l1_portal"]
	}

	pUUID, reuse, err := m.etl.CreateDataPipeline(cfg)
	if err != nil {
		return core.NilSUUID(), err
	}

	reg, err := m.etl.GetRegister(cfg.DataType)
	if err != nil {
		return core.NilSUUID(), err
	}

	logger.Info("Created etl pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	deployCfg := &invariant.DeployConfig{
		PUUID:     pUUID,
		InvType:   invCfg.Type,
		InvParams: invCfg.Params,
		Network:   cfg.Network,
		Register:  reg,
	}

	sUUID, err := m.eng.DeployInvariantSession(deployCfg)
	if err != nil {
		return core.NilSUUID(), err
	}
	logger.Info("Deployed invariant session", zap.String(core.SUUIDKey, sUUID.String()))

	err = m.alrt.AddInvariantSession(sUUID, invCfg.AlertDest)
	if err != nil {
		return core.NilSUUID(), err
	}

	if reuse { // If the pipeline was reused, we don't need to run it again
		return sUUID, nil
	}

	if err = m.etl.RunPipeline(pUUID); err != nil {
		return core.NilSUUID(), err
	}

	return sUUID, nil
}
