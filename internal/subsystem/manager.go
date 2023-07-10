//go:generate mockgen -package mocks --destination ../mocks/subsystem.go --mock_names Manager=SubManager . Manager

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
	BuildDeployCfg(pConfig *core.PipelineConfig, sConfig *core.SessionConfig) (*invariant.DeployConfig, error)
	BuildPipelineCfg(params *models.InvRequestParams) (*core.PipelineConfig, error)
	RunInvSession(cfg *invariant.DeployConfig) (core.SUUID, error)
	// Orchestration
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
func (m *manager) Shutdown() error {
	// 1. Shutdown ETL subsystem
	if err := m.etl.Shutdown(); err != nil {
		return err
	}

	// 2. Shutdown Engine subsystem
	if err := m.eng.Shutdown(); err != nil {
		return err
	}

	// 3. Shutdown Alert subsystem
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

// BuildDeployCfg ... Builds a deploy config provided a pipeline & session config
func (m *manager) BuildDeployCfg(pConfig *core.PipelineConfig,
	sConfig *core.SessionConfig) (*invariant.DeployConfig, error) {
	// 1. Fetch state key using risk engine input register type
	sk, stateful, err := m.etl.GetStateKey(pConfig.DataType)
	if err != nil {
		return nil, err
	}

	// 2. Create data pipeline
	pUUID, reuse, err := m.etl.CreateDataPipeline(pConfig)
	if err != nil {
		return nil, err
	}

	logging.WithContext(m.ctx).
		Info("Created etl pipeline", zap.String(core.PUUIDKey, pUUID.String()))

	// 3. Create a deploy config
	return &invariant.DeployConfig{
		PUUID:     pUUID,
		Reuse:     reuse,
		InvType:   sConfig.Type,
		InvParams: sConfig.Params,
		Network:   pConfig.Network,
		Stateful:  stateful,
		StateKey:  sk,
		AlertDest: sConfig.AlertDest,
	}, nil
}

// RunInvSession ... Runs an invariant session
func (m *manager) RunInvSession(cfg *invariant.DeployConfig) (core.SUUID, error) {
	// 1. Deploy invariant session to risk engine
	sUUID, err := m.eng.DeployInvariantSession(cfg)
	if err != nil {
		return core.NilSUUID(), err
	}
	logging.WithContext(m.ctx).
		Info("Deployed invariant session to risk engine", zap.String(core.SUUIDKey, sUUID.String()))

	// 2. Add session to alert manager
	err = m.alrt.AddSession(sUUID, cfg.AlertDest)
	if err != nil {
		return core.NilSUUID(), err
	}

	// 3. Run pipeline if not reused
	if cfg.Reuse {
		return sUUID, nil
	}

	if err = m.etl.RunPipeline(cfg.PUUID); err != nil { // Spin-up pipeline components
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
		PipelineType: params.PipelineType(),
		ClientConfig: &core.ClientConfig{
			Network:      params.NetworkType(),
			PollInterval: pollInterval,
			StartHeight:  params.StartHeight,
			EndHeight:    params.EndHeight,
		},
	}, nil
}
