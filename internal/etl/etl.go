//go:generate mockgen -package mocks --destination ../mocks/etl.go . ETL

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

type ETL interface {
	CreateProcess(cc *core.ClientConfig, id core.ProcessID, PathID core.PathID,
		dt *core.DataTopic) (process.Process, error)
	GetStateKey(rt core.TopicType) (*core.StateKey, bool, error)
	GetBlockHeight(id core.PathID) (*big.Int, error)
	CreateProcessPath(cfg *core.PathConfig) (core.PathID, bool, error)
	Run(id core.PathID) error
	ActiveCount() int

	core.Subsystem
}
type etl struct {
	ctx    context.Context
	cancel context.CancelFunc

	analyzer Analyzer
	dag      *Graph
	store    *Store
	metrics  metrics.Metricer

	egress chan core.HeuristicInput

	registry *registry.Registry
	wg       sync.WaitGroup
}

// New ... Initializer
func New(ctx context.Context, analyzer Analyzer, r *registry.Registry,
	store *Store, dag *Graph, eo chan core.HeuristicInput) ETL {
	ctx, cancel := context.WithCancel(ctx)
	stats := metrics.WithContext(ctx)
	return &etl{
		analyzer: analyzer,
		ctx:      ctx,
		cancel:   cancel,
		dag:      dag,
		store:    store,
		registry: r,
		egress:   eo,
		metrics:  stats,
		wg:       sync.WaitGroup{},
	}
}

// GetDataTopic ... Returns a data register for a given register type
func (etl *etl) GetDataTopic(rt core.TopicType) (*core.DataTopic, error) {
	return etl.registry.GetDataTopic(rt)
}

func (etl *etl) CreateProcessPath(cfg *core.PathConfig) (core.PathID, bool, error) {
	// NOTE - If some of these early sub-system operations succeed but lower function
	// code logic fails, then some rollback will need be triggered to undo prior applied state operations
	logger := logging.WithContext(etl.ctx)

	depPath, err := etl.registry.TopicPath(cfg.DataType)
	if err != nil {
		return core.PathID{}, false, err
	}

	id := depPath.GeneratePathID(cfg.PathType, cfg.Network)

	processes, err := etl.topicPath(cfg, id, depPath)
	if err != nil {
		return core.PathID{}, false, err
	}

	logger.Debug("Constructing path",
		zap.String(logging.Path, id.String()))

	path, err := NewPath(cfg, id, processes)
	if err != nil {
		return core.PathID{}, false, err
	}

	mergeID, err := etl.getMergePath(id, path)
	if err != nil {
		return core.PathID{}, false, err
	}

	nilID := core.PathID{}
	if mergeID != nilID { // An existing path can be reused
		return mergeID, true, nil
	}

	// Bind communication route between path and risk engine
	if err := path.AddEngineRelay(etl.egress); err != nil {
		return core.PathID{}, false, err
	}

	// Add path object to the store
	etl.store.AddPath(id, path)

	return id, false, nil
}

// RunPath ...
func (etl *etl) Run(id core.PathID) error {
	// 1. Get path from store
	path, err := etl.store.GetPathByID(id)
	if err != nil {
		return err
	}

	// 2. Add path processes to graph
	if err := etl.dag.AddMany(path.Processes()); err != nil {
		return err
	}

	logging.WithContext(etl.ctx).Info("Running path",
		zap.String(logging.Path, id.String()))

	// 3. Run processes
	path.Run(&etl.wg)
	etl.metrics.IncActivePaths(id.NetworkType())
	return nil
}

// EventLoop ... Driver ran as separate go routine
func (etl *etl) EventLoop() error {
	logger := logging.WithContext(etl.ctx)

	for {
		<-etl.ctx.Done()
		logger.Info("Shutting down ETL")
		return nil
	}
}

// Shutdown ... Shuts down all paths
func (etl *etl) Shutdown() error {
	etl.cancel()
	logger := logging.WithContext(etl.ctx)

	for _, p := range etl.store.Paths() {
		logger.Info("Shutting down path",
			zap.String(logging.Path, p.UUID().String()))

		if err := p.Close(); err != nil {
			logger.Error("Failed to close path",
				zap.String(logging.Path, p.UUID().String()))
			return err
		}
		etl.metrics.DecActivePaths(p.UUID().NetworkType())
	}
	logger.Debug("Waiting for all process routines to end")
	etl.wg.Wait()

	return nil
}

// ActiveCount ... Returns the number of active paths
func (etl *etl) ActiveCount() int {
	return etl.store.ActiveCount()
}

func (etl *etl) topicPath(cfg *core.PathConfig, pathID core.PathID,
	depPath core.TopicPath) ([]process.Process, error) {
	processes := make([]process.Process, 0)

	for _, register := range depPath.Path {
		id := core.MakeProcessID(cfg.PathType, register.ProcessType, register.DataType, cfg.Network)

		p, err := etl.CreateProcess(cfg.ClientConfig, id, pathID, register)
		if err != nil {
			return []process.Process{}, err
		}

		processes = append(processes, p)
	}

	return processes, nil
}

// getMergePath ... Returns a path UUID if a merging opportunity exists
func (etl *etl) getMergePath(id core.PathID, path Path) (core.PathID, error) {
	paths := etl.store.GetExistingPaths(id)

	for _, id := range paths {
		p, err := etl.store.GetPathByID(id)
		if err != nil {
			return core.PathID{}, err
		}

		if etl.analyzer.Mergable(path, p) { // Deploy heuristics to existing paths instead
			// This is a bit hacky since we aren't actually merging the paths
			return p.UUID(), nil
		}
	}

	return core.PathID{}, nil
}

func (etl *etl) CreateProcess(cc *core.ClientConfig, id core.ProcessID, pathID core.PathID,
	dt *core.DataTopic) (process.Process, error) {
	logging.WithContext(etl.ctx).Debug("constructing process",
		zap.String("type", dt.ProcessType.String()),
		zap.String("register_type", dt.DataType.String()))

	// embed options to avoid constructor boilerplate
	opts := []process.Option{process.WithID(id), process.WithPathID(pathID)}

	if dt.Stateful() {
		// Propagate state key to process so that it can be used
		// by the process's definition logic
		sk := dt.StateKey()
		err := sk.SetPathID(pathID)
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

		return init(etl.ctx, cc, opts...)

	case core.Subscribe:
		init, success := dt.Constructor.(process.Constructor)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Subscribe.String()))
		}

		return init(etl.ctx, cc, opts...)

	default:
		return nil, fmt.Errorf(unknownCompType, dt.ProcessType.String())
	}
}

// GetStateKey ... Returns a state key provided a register type
func (etl *etl) GetStateKey(rt core.TopicType) (*core.StateKey, bool, error) {
	dr, err := etl.registry.GetDataTopic(rt)
	if err != nil {
		return nil, false, err
	}

	if dr.Stateful() {
		return dr.StateKey(), true, nil
	}

	return nil, false, nil
}

func (etl *etl) GetBlockHeight(id core.PathID) (*big.Int, error) {
	path, err := etl.store.GetPathByID(id)
	if err != nil {
		return nil, err
	}

	height, err := path.BlockHeight()
	if err != nil {
		return nil, err
	}

	return height, nil
}
