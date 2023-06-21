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
	"go.uber.org/zap"
)

// Manager ... ETL manager interface
type Manager interface {
	GetRegister(rt core.RegisterType) (*core.DataRegister, error)
	CreateDataPipeline(cfg *core.PipelineConfig) (core.PUUID, error)
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

	engOutgress chan core.InvariantInput

	registry registry.Registry
	wg       sync.WaitGroup
}

// NewManager ... Initializer
func NewManager(ctx context.Context, analyzer Analyzer, cRegistry registry.Registry,
	store EtlStore, dag ComponentGraph,
	eo chan core.InvariantInput) Manager {
	m := &etlManager{
		analyzer:    analyzer,
		ctx:         ctx,
		dag:         dag,
		store:       store,
		registry:    cRegistry,
		engOutgress: eo,
		wg:          sync.WaitGroup{},
	}

	return m
}

// GetRegister ... Returns a data register for a given register type
func (em *etlManager) GetRegister(rt core.RegisterType) (*core.DataRegister, error) {
	return em.registry.GetRegister(rt)
}

// CreateDataPipeline ... Creates an ETL data pipeline provided a pipeline configuration
func (em *etlManager) CreateDataPipeline(cfg *core.PipelineConfig) (core.PUUID, error) {
	// NOTE - If some of these early sub-system operations succeed but lower function
	// code logic fails, then some rollback will need be triggered to undo prior applied state operations
	logger := logging.WithContext(em.ctx)

	depPath, err := em.registry.GetDependencyPath(cfg.DataType)
	if err != nil {
		return core.NilPUUID(), err
	}

	pUUID := depPath.GeneratePUUID(cfg.PipelineType, cfg.Network)

	components, err := em.getComponents(cfg, depPath)
	if err != nil {
		return core.NilPUUID(), err
	}

	logger.Debug("constructing pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	pipeline, err := NewPipeline(cfg, pUUID, components)
	if err != nil {
		return core.NilPUUID(), err
	}

	mPUUID, err := em.getMergeUUID(pUUID, pipeline)
	if err != nil {
		return core.NilPUUID(), err
	}

	if mPUUID != core.NilPUUID() { // Pipeline is mergable
		return pUUID, nil
	}

	// Bind communication route between pipeline and risk engine
	if err := pipeline.AddEngineRelay(em.engOutgress); err != nil {
		return core.NilPUUID(), err
	}

	if err := em.dag.AddComponents(pipeline.Components()); err != nil {
		return core.NilPUUID(), err
	}

	em.store.AddPipeline(pUUID, pipeline)

	if len(components) == 1 {
		return pUUID, nil
	}

	return pUUID, nil
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
	logger := logging.WithContext(em.ctx)

	for _, pl := range em.store.GetAllPipelines() {
		logger.Info("Shuting down pipeline",
			zap.String(core.PUUIDKey, pl.UUID().String()))

		if err := pl.Close(); err != nil {
			logger.Error("Failed to close pipeline",
				zap.String(core.PUUIDKey, pl.UUID().String()))
			return err
		}
	}
	logger.Debug("Waiting for all component routines to end")
	em.wg.Wait()
	em.cancel()

	return nil
}

// getComponents ... Returns all components provided a slice of register definitions
func (em *etlManager) getComponents(cfg *core.PipelineConfig,
	depPath core.RegisterDependencyPath) ([]component.Component, error) {
	components := make([]component.Component, 0)
	// prevUUID := core.NilCUUID()

	for _, register := range depPath.Path {
		cUUID := core.MakeCUUID(cfg.PipelineType, register.ComponentType, register.DataType, cfg.Network)

		c, err := inferComponent(em.ctx, cfg.ClientConfig, cUUID, register)
		if err != nil {
			return []component.Component{}, err
		}

		components = append(components, c)
	}

	return components, nil
}

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

// inferComponent ... Constructs a component provided a data register definition
func inferComponent(ctx context.Context, cc *core.ClientConfig, cUUID core.CUUID,
	register *core.DataRegister) (component.Component, error) {
	logging.WithContext(ctx).Debug("constructing component",
		zap.String("type", register.ComponentType.String()),
		zap.String("outdata_type", register.DataType.String()))

	opts := []component.Option{component.WithCUUID(cUUID)}

	if register.Stateful() {
		opts = append(opts, component.WithStateKey(register.StateKey))
	}

	switch register.ComponentType {
	case core.Oracle:
		init, success := register.ComponentConstructor.(component.OracleConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Oracle.String()))
		}

		return init(ctx, cc, opts...)

	case core.Pipe:
		init, success := register.ComponentConstructor.(component.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Pipe.String()))
		}

		return init(ctx, cc, opts...)

	case core.Aggregator:
		return nil, fmt.Errorf("aggregator component has yet to be implemented")

	default:
		return nil, fmt.Errorf(unknownCompType, register.ComponentType.String())
	}
}
