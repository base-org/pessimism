package subsystem

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Config ... Used to store necessary API service config values
type Config struct {
	L1PollInterval int
	L2PollInterval int
}

// GetPollInterval ... Returns config poll-interval for network type
func (cfg *Config) GetPollInterval(n core.Network) (time.Duration, error) {
	switch n {
	case core.Layer1:
		return time.Duration(cfg.L1PollInterval), nil

	case core.Layer2:
		return time.Duration(cfg.L2PollInterval), nil

	default:
		return 0, fmt.Errorf("could not find endpoint for network %s", n.String())
	}
}

// Manager ... Subsystem manager interface
type Manager interface {
	BuildPipelineCfg(params *models.InvRequestParams) (*core.PipelineConfig, error)
	RunInvSession(pConfig *core.PipelineConfig, sConfig *core.SessionConfig) (core.SUUID, error)
	StartEventRoutines(ctx context.Context)
	Shutdown() error
}

// manager ... Subsystem manager struct
type manager struct {
	cfg *Config
	ctx context.Context

	etl  pipeline.Manager
	eng  engine.Manager
	alrt alert.Manager

	*sync.WaitGroup
}

// NewManager ... Initializer for the subsystem manager
func NewManager(ctx context.Context, cfg *Config, etl pipeline.Manager, eng engine.Manager,
	alrt alert.Manager,
) Manager {
	return &manager{
		cfg:       cfg,
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

func (m *manager) RunInvSession(pConfig *core.PipelineConfig, sConfig *core.SessionConfig) (core.SUUID, error) {
	logger := logging.WithContext(m.ctx)

	sk, stateful, err := m.etl.GetStateKey(pConfig.DataType)
	if err != nil {
		return core.NilSUUID(), err
	}

	pUUID, reuse, err := m.etl.CreateDataPipeline(pConfig)

	logger.Info("Created etl pipeline", zap.String(core.PUUIDKey, pUUID.String()))

	deployCfg := &invariant.DeployConfig{
		PUUID:     pUUID,
		InvType:   sConfig.Type,
		InvParams: sConfig.Params,
		Network:   pConfig.Network,
		Stateful:  stateful,
		StateKey:  sk,
	}

	sUUID, err := m.eng.DeployInvariantSession(deployCfg)
	if err != nil {
		return core.NilSUUID(), err
	}
	logger.Info("Deployed invariant session to risk engine", zap.String(core.SUUIDKey, sUUID.String()))

	err = m.alrt.AddInvariantSession(sUUID, sConfig.AlertDest)
	if err != nil {
		return core.NilSUUID(), err
	}

	if reuse { // If the pipeline was reused, we don't need to run it again
		return sUUID, nil
	}

	if err = m.etl.RunPipeline(pUUID); err != nil { // Spinup pipeline components
		return core.NilSUUID(), err
	}

	return sUUID, nil

}

// BuildPipelineCfg ... Builds a pipeline config provided a set of invariant request params
func (m *manager) BuildPipelineCfg(params *models.InvRequestParams) (*core.PipelineConfig, error) {
	inType, err := m.eng.GetInputType(params.InvariantType())
	if err != nil {
		return nil, err
	}

	pollInterval, err := m.cfg.GetPollInterval(params.NetworkType())
	if err != nil {
		return nil, err
	}

	return &core.PipelineConfig{
		Network:      params.NetworkType(),
		DataType:     inType,
		PipelineType: params.PiplineType(),
		ClientConfig: &core.ClientConfig{
			Network:      params.NetworkType(),
			PollInterval: pollInterval,
			NumOfRetries: 3,
			StartHeight:  params.StartHeight,
			EndHeight:    params.EndHeight,
		},
	}, nil

}
