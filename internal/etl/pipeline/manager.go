//go:generate mockgen -package mocks --destination ../../mocks/etl_manager.go --mock_names Manager=EtlManager . Manager

package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"

	"go.uber.org/zap"
)

// Manager ... ETL manager interface
type Manager interface {
	InferComponent(cc *core.ClientConfig, cUUID core.CUUID, pUUID core.PUUID,
		register *core.DataRegister) (component.Component, error)
	GetRegister(rt core.RegisterType) (*core.DataRegister, error)
	CreateDataPipeline(cfg *core.PipelineConfig) (core.PUUID, bool, error)
	RunPipeline(pID core.PUUID) error

	core.Subsystem
}

// etlManager ... ETL manager
type etlManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	analyzer Analyzer
	dag      ComponentGraph
	store    EtlStore
	metrics  metrics.Metricer

	engOutgress chan core.InvariantInput

	registry registry.Registry
	wg       sync.WaitGroup
}

// NewManager ... Initializer
func NewManager(ctx context.Context, analyzer Analyzer, cRegistry registry.Registry,
	store EtlStore, dag ComponentGraph,
	eo chan core.InvariantInput) Manager {
	ctx, cancel := context.WithCancel(ctx)
	stats := metrics.WithContext(ctx)

	m := &etlManager{
		analyzer:    analyzer,
		ctx:         ctx,
		cancel:      cancel,
		dag:         dag,
		store:       store,
		registry:    cRegistry,
		engOutgress: eo,
		metrics:     stats,
		wg:          sync.WaitGroup{},
	}

	return m
}

// GetRegister ... Returns a data register for a given register type
func (em *etlManager) GetRegister(rt core.RegisterType) (*core.DataRegister, error) {
	return em.registry.GetRegister(rt)
}

// CreateDataPipeline ... Creates an ETL data pipeline provided a pipeline configuration
func (em *etlManager) CreateDataPipeline(cfg *core.PipelineConfig) (core.PUUID, bool, error) {
	// NOTE - If some of these early sub-system operations succeed but lower function
	// code logic fails, then some rollback will need be triggered to undo prior applied state operations
	logger := logging.WithContext(em.ctx)

	depPath, err := em.registry.GetDependencyPath(cfg.DataType)
	if err != nil {
		return core.NilPUUID(), false, err
	}

	pUUID := depPath.GeneratePUUID(cfg.PipelineType, cfg.Network)

	components, err := em.getComponents(cfg, pUUID, depPath)
	if err != nil {
		return core.NilPUUID(), false, err
	}

	logger.Debug("constructing pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	pipeline, err := NewPipeline(cfg, pUUID, components)
	if err != nil {
		return core.NilPUUID(), false, err
	}

	mPUUID, err := em.getMergeUUID(pUUID, pipeline)
	if err != nil {
		return core.NilPUUID(), false, err
	}

	if mPUUID != core.NilPUUID() { // Pipeline is mergable
		return mPUUID, true, nil
	}

	// Bind communication route between pipeline and risk engine
	if err := pipeline.AddEngineRelay(em.engOutgress); err != nil {
		return core.NilPUUID(), false, err
	}

	if err := em.dag.AddComponents(pipeline.Components()); err != nil {
		return core.NilPUUID(), false, err
	}

	em.store.AddPipeline(pUUID, pipeline)

	if len(components) == 1 {
		return pUUID, false, nil
	}

	em.metrics.IncActivePipelines()

	return pUUID, false, nil
}

// RunPipeline ... Runs pipeline session for some provided pUUID
func (em *etlManager) RunPipeline(pUUID core.PUUID) error {
	pipeline, err := em.store.GetPipelineFromPUUID(pUUID)
	if err != nil {
		return err
	}

	logging.WithContext(em.ctx).Info("Running pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	return pipeline.RunPipeline(&em.wg)
}

// EventLoop ... Driver ran as separate go routine
func (em *etlManager) EventLoop() error {
	logger := logging.WithContext(em.ctx)

	for {
		<-em.ctx.Done()
		logger.Info("Receieved shutdown request")
		return nil
	}
}

// Shutdown ... Shuts down all pipelines
func (em *etlManager) Shutdown() error {
	em.cancel()
	logger := logging.WithContext(em.ctx)

	for _, pl := range em.store.GetAllPipelines() {
		logger.Info("Shutting down pipeline",
			zap.String(core.PUUIDKey, pl.UUID().String()))

		if err := pl.Close(); err != nil {
			logger.Error("Failed to close pipeline",
				zap.String(core.PUUIDKey, pl.UUID().String()))
			return err
		}
		em.metrics.DecActivePipelines()
	}
	logger.Debug("Waiting for all component routines to end")
	em.wg.Wait()

	return nil
}

// getComponents ... Returns all components provided a slice of register definitions
func (em *etlManager) getComponents(cfg *core.PipelineConfig, pUUID core.PUUID,
	depPath core.RegisterDependencyPath) ([]component.Component, error) {
	components := make([]component.Component, 0)

	for _, register := range depPath.Path {
		cUUID := core.MakeCUUID(cfg.PipelineType, register.ComponentType, register.DataType, cfg.Network)

		c, err := em.InferComponent(cfg.ClientConfig, cUUID, pUUID, register)
		if err != nil {
			return []component.Component{}, err
		}

		components = append(components, c)
	}

	return components, nil
}

// getMergeUUID ... Returns a pipeline UUID if a mergable pipeline exists
func (em *etlManager) getMergeUUID(pUUID core.PUUID, pipeline Pipeline) (core.PUUID, error) {
	pipelines := em.store.GetExistingPipelinesByPID(pUUID.PID)

	for _, pl := range pipelines {
		p, err := em.store.GetPipelineFromPUUID(pl)
		if err != nil {
			return core.NilPUUID(), err
		}

		if em.analyzer.Mergable(pipeline, p) { // Deploy invariants to existing pipelines instead
			// This is a bit hacky since we aren't actually merging the pipelines
			return p.UUID(), nil
		}
	}

	return core.NilPUUID(), nil
}

// InferComponent ... Constructs a component provided a data register definition
func (em *etlManager) InferComponent(cc *core.ClientConfig, cUUID core.CUUID, pUUID core.PUUID,
	register *core.DataRegister) (component.Component, error) {
	logging.WithContext(em.ctx).Debug("constructing component",
		zap.String("type", register.ComponentType.String()),
		zap.String("outdata_type", register.DataType.String()))

	opts := []component.Option{component.WithCUUID(cUUID), component.WithPUUID(pUUID)}

	if register.Stateful() {
		// Propagate state key to component so that it can be used
		// by the component's definition logic
		sk := register.StateKey()
		err := sk.SetPUUID(pUUID)
		if err != nil {
			return nil, err
		}

		opts = append(opts, component.WithStateKey(sk))
	}

	switch register.ComponentType {
	case core.Oracle:
		init, success := register.ComponentConstructor.(component.OracleConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Oracle.String()))
		}

		return init(em.ctx, cc, opts...)

	case core.Pipe:
		init, success := register.ComponentConstructor.(component.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Pipe.String()))
		}

		return init(em.ctx, cc, opts...)

	case core.Aggregator:
		return nil, fmt.Errorf(noAggregatorErr)

	default:
		return nil, fmt.Errorf(unknownCompType, register.ComponentType.String())
	}
}
