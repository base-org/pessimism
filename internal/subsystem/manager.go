//go:generate mockgen -package mocks --destination ../mocks/subsystem.go --mock_names Subsystem=SubManager . Subsystem

package subsystem

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/etl"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

// Config ... Used to store necessary API service config values
type Config struct {
	MaxPathCount   int
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
		return 0, fmt.Errorf(networkNotFoundErr, n.String())
	}
}

type Subsystem interface {
	BuildDeployCfg(pConfig *core.PathConfig, sConfig *core.SessionConfig) (*heuristic.DeployConfig, error)
	BuildPathCfg(params *models.SessionRequestParams) (*core.PathConfig, error)
	RunHeuristic(cfg *heuristic.DeployConfig) (core.UUID, error)
	// Orchestration
	StartEventRoutines(ctx context.Context)
	Shutdown() error
}

// manager ... Subsystem manager struct
type Manager struct {
	cfg *Config
	ctx context.Context

	etl   etl.ETL
	eng   engine.Manager
	alert alert.Manager
	stats metrics.Metricer

	*sync.WaitGroup
}

// NewManager ... Initializer for the subsystem manager
func NewManager(ctx context.Context, cfg *Config, etl etl.ETL, eng engine.Manager,
	a alert.Manager,
) *Manager {
	return &Manager{
		cfg:       cfg,
		ctx:       ctx,
		etl:       etl,
		eng:       eng,
		alert:     a,
		stats:     metrics.WithContext(ctx),
		WaitGroup: &sync.WaitGroup{},
	}
}

// Shutdown ... Shuts down all subsystems in primary data flow order
func (m *Manager) Shutdown() error {
	// 1. Shutdown ETL subsystem
	if err := m.etl.Shutdown(); err != nil {
		return err
	}

	// 2. Shutdown Engine subsystem
	if err := m.eng.Shutdown(); err != nil {
		return err
	}

	// 3. Shutdown Alert subsystem
	return m.alert.Shutdown()
}

// StartEventRoutines ... Starts the event loop routines for the subsystems
func (m *Manager) StartEventRoutines(ctx context.Context) {
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

		if err := m.alert.EventLoop(); err != nil {
			logger.Error("alert manager event loop error", zap.Error(err))
		}
	}()

	m.Add(1)
	go func() { // ETL driver thread
		defer m.Done()

		if err := m.etl.EventLoop(); err != nil {
			logger.Error("ETL manager event loop error", zap.Error(err))
		}
	}()
}

// BuildDeployCfg ... Builds a deploy config provided a path & session config
func (m *Manager) BuildDeployCfg(pConfig *core.PathConfig,
	sConfig *core.SessionConfig) (*heuristic.DeployConfig, error) {
	// 1. Fetch state key using risk engine input register type
	sk, stateful, err := m.etl.GetStateKey(pConfig.DataType)
	if err != nil {
		return nil, err
	}

	// 2. Create data path
	id, reuse, err := m.etl.CreateProcessPath(pConfig)
	if err != nil {
		return nil, err
	}

	logging.WithContext(m.ctx).
		Info("Created etl path", zap.String(logging.Path, id.String()))

	// 3. Create a deploy config
	return &heuristic.DeployConfig{
		PathID:         id,
		Reuse:          reuse,
		HeuristicType:  sConfig.Type,
		Params:         sConfig.Params,
		Network:        pConfig.Network,
		Stateful:       stateful,
		StateKey:       sk,
		AlertingPolicy: sConfig.AlertPolicy,
	}, nil
}

// RunHeuristic ... Runs a heuristic session
func (m *Manager) RunHeuristic(cfg *heuristic.DeployConfig) (core.UUID, error) {
	// 1. Verify that path constraints are met
	// NOTE - Consider introducing a config validation step or module
	if !cfg.Reuse && m.etlLimitReached() {
		return core.UUID{}, fmt.Errorf(maxPathErr, m.cfg.MaxPathCount)
	}

	// 2. Deploy heuristic session to risk engine
	id, err := m.eng.DeployHeuristic(cfg)
	if err != nil {
		return core.UUID{}, err
	}
	logging.WithContext(m.ctx).
		Info("Deployed heuristic session to risk engine", zap.String(logging.UUID, id.ShortString()))

	// 3. Add session to alert manager
	err = m.alert.AddSession(id, cfg.AlertingPolicy)
	if err != nil {
		return core.UUID{}, err
	}

	// 4. Run path if not reused
	if cfg.Reuse {
		return id, nil
	}

	if err = m.etl.Run(cfg.PathID); err != nil { // Spin-up path processes
		return core.UUID{}, err
	}

	return id, nil
}

// BuildPathCfg ... Builds a path config provided a set of heuristic request params
func (m *Manager) BuildPathCfg(params *models.SessionRequestParams) (*core.PathConfig, error) {
	inType, err := m.eng.GetInputType(params.Heuristic())
	if err != nil {
		return nil, err
	}

	pollInterval, err := m.cfg.GetPollInterval(params.NetworkType())
	if err != nil {
		return nil, err
	}

	return &core.PathConfig{
		Network:  params.NetworkType(),
		DataType: inType,
		PathType: core.Live,
		ClientConfig: &core.ClientConfig{
			Network:      params.NetworkType(),
			PollInterval: pollInterval,
			StartHeight:  params.StartHeight,
			EndHeight:    params.EndHeight,
		},
	}, nil
}

// etlLimitReached ... Returns true if the ETL path count is at or above the max
func (m *Manager) etlLimitReached() bool {
	return m.etl.ActiveCount() >= m.cfg.MaxPathCount
}

func (m *Manager) PathHeight(id core.PathID) (*big.Int, error) {
	return m.etl.GetBlockHeight(id)
}
