//go:generate mockgen -package mocks --destination ../../mocks/etl_manager.go --mock_names Manager=EtlManager . Manager

package etl

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/etl/registry"

	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"

	"go.uber.org/zap"
)

// Manager ... ETL manager interface
type Manager interface {
	CreateProcess(cc *core.ClientConfig, cUUID core.ProcessID, PathID core.PathID,
		dt *core.DataTopic) (process.Process, error)
	GetStateKey(rt core.TopicType) (*core.StateKey, bool, error)
	GetHeightAtPath(id core.PathID) (*big.Int, error)
	CreateProcessPath(cfg *core.PathConfig) (core.PathID, bool, error)
	StartPath(pID core.PathID) error
	ActiveCount() int

	core.Subsystem
}

// etlManager ... ETL manager
type etlManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	analyzer Analyzer
	dag      *Graph
	store    EtlStore
	metrics  metrics.Metricer

	egress chan core.HeuristicInput

	registry registry.Registry
	wg       sync.WaitGroup
}

// NewManager ... Initializer
func NewManager(ctx context.Context, analyzer Analyzer, cRegistry registry.Registry,
	store EtlStore, dag *Graph, eo chan core.HeuristicInput) Manager {
	ctx, cancel := context.WithCancel(ctx)
	stats := metrics.WithContext(ctx)

	m := &etlManager{
		analyzer: analyzer,
		ctx:      ctx,
		cancel:   cancel,
		dag:      dag,
		store:    store,
		registry: cRegistry,
		egress:   eo,
		metrics:  stats,
		wg:       sync.WaitGroup{},
	}

	return m
}

// GetDataTopic ... Returns a data register for a given register type
func (em *etlManager) GetDataTopic(rt core.TopicType) (*core.DataTopic, error) {
	return em.registry.GetDataTopic(rt)
}

func (em *etlManager) CreateProcessPath(cfg *core.PathConfig) (core.PathID, bool, error) {
	// NOTE - If some of these early sub-system operations succeed but lower function
	// code logic fails, then some rollback will need be triggered to undo prior applied state operations
	logger := logging.WithContext(em.ctx)

	depPath, err := em.registry.GetDependencyPath(cfg.DataType)
	if err != nil {
		return core.PathID{}, false, err
	}

	id := depPath.GeneratePathID(cfg.PathType, cfg.Network)

	components, err := em.getComponents(cfg, id, depPath)
	if err != nil {
		return core.PathID{}, false, err
	}

	logger.Debug("Constructing pipeline",
		zap.String(logging.PathIDKey, id.String()))

	path, err := NewPath(cfg, id, components)
	if err != nil {
		return core.PathID{}, false, err
	}

	mergeID, err := em.getMergePath(id, path)
	if err != nil {
		return core.PathID{}, false, err
	}

	nilID := core.PathID{}
	if mergeID != nilID { // An existing path can be reused
		return mergeID, true, nil
	}

	// Bind communication route between pipeline and risk engine
	if err := path.AddEngineRelay(em.egress); err != nil {
		return core.PathID{}, false, err
	}

	// Add pipeline object to the store
	em.store.AddPath(id, path)

	return id, false, nil
}

// StartPath ...
func (em *etlManager) StartPath(id core.PathID) error {
	// 1. Get pipeline from store
	path, err := em.store.GetPathByID(id)
	if err != nil {
		return err
	}

	// 2. Add pipeline components to the component graph
	if err := em.dag.AddMany(path.Processes()); err != nil {
		return err
	}

	logging.WithContext(em.ctx).Info("Running pipeline",
		zap.String(logging.PathIDKey, id.String()))

	// 3. Run pipeline
	path.Run(&em.wg)

	// Pipeline successfully created, increment for type and network
	em.metrics.IncActivePipelines(id.PathType(), id.NetworkType())
	return nil
}

// EventLoop ... Driver ran as separate go routine
func (em *etlManager) EventLoop() error {
	logger := logging.WithContext(em.ctx)

	for {
		<-em.ctx.Done()
		logger.Info("Received shutdown request")
		return nil
	}
}

// Shutdown ... Shuts down all pipelines
func (em *etlManager) Shutdown() error {
	em.cancel()
	logger := logging.WithContext(em.ctx)

	for _, pl := range em.store.Paths() {
		logger.Info("Shutting down pipeline",
			zap.String(logging.PathIDKey, pl.UUID().String()))

		if err := pl.Close(); err != nil {
			logger.Error("Failed to close pipeline",
				zap.String(logging.PathIDKey, pl.UUID().String()))
			return err
		}
		em.metrics.DecActivePipelines(pl.UUID().PathType(), pl.UUID().NetworkType())
	}
	logger.Debug("Waiting for all component routines to end")
	em.wg.Wait()

	return nil
}

// ActiveCount ... Returns the number of active pipelines
func (em *etlManager) ActiveCount() int {
	return em.store.ActiveCount()
}

// getComponents ... Returns all components provided a slice of register definitions
func (em *etlManager) getComponents(cfg *core.PathConfig, PathID core.PathID,
	depPath core.RegisterDependencyPath) ([]process.Process, error) {
	components := make([]process.Process, 0)

	for _, register := range depPath.Path {
		id := core.MakeProcessID(cfg.PathType, register.ProcessType, register.DataType, cfg.Network)

		c, err := em.CreateProcess(cfg.ClientConfig, id, PathID, register)
		if err != nil {
			return []process.Process{}, err
		}

		components = append(components, c)
	}

	return components, nil
}

// getMergePath ... Returns a pipeline UUID if a merging opportunity exists
func (em *etlManager) getMergePath(id core.PathID, path Path) (core.PathID, error) {
	paths := em.store.GetExistingPaths(id)

	for _, id := range paths {
		p, err := em.store.GetPathByID(id)
		if err != nil {
			return core.PathID{}, err
		}

		if em.analyzer.Mergable(path, p) { // Deploy heuristics to existing paths instead
			// This is a bit hacky since we aren't actually merging the paths
			return p.UUID(), nil
		}
	}

	return core.PathID{}, nil
}

func (em *etlManager) CreateProcess(cc *core.ClientConfig, cUUID core.ProcessID, PathID core.PathID,
	dt *core.DataTopic) (process.Process, error) {
	logging.WithContext(em.ctx).Debug("constructing component",
		zap.String("type", dt.ProcessType.String()),
		zap.String("register_type", dt.DataType.String()))

	// Embed options to avoid constructor boilerplate
	opts := []process.Option{process.WithID(cUUID), process.WithPathID(PathID)}

	if dt.Stateful() {
		// Propagate state key to component so that it can be used
		// by the component's definition logic
		sk := dt.StateKey()
		err := sk.SetPathID(PathID)
		if err != nil {
			return nil, err
		}

		opts = append(opts, process.WithStateKey(sk))
	}

	switch dt.ProcessType {
	case core.Read:
		init, success := dt.Constructor.(process.Constructor)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Read.String()))
		}

		return init(em.ctx, cc, opts...)

	case core.Subscribe:
		init, success := dt.Constructor.(process.Constructor)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Subscribe.String()))
		}

		return init(em.ctx, cc, opts...)

	default:
		return nil, fmt.Errorf(unknownCompType, dt.ProcessType.String())
	}
}

// GetStateKey ... Returns a state key provided a register type
func (em *etlManager) GetStateKey(rt core.TopicType) (*core.StateKey, bool, error) {
	dr, err := em.registry.GetDataTopic(rt)
	if err != nil {
		return nil, false, err
	}

	if dr.Stateful() {
		return dr.StateKey(), true, nil
	}

	return nil, false, nil
}

func (em *etlManager) GetHeightAtPath(id core.PathID) (*big.Int, error) {
	path, err := em.store.GetPathByID(id)
	if err != nil {
		return nil, err
	}

	height, err := path.BlockHeight()
	if err != nil {
		return nil, err
	}

	return height, nil
}
