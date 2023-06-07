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
	CreateDataPipeline(cfg *core.PipelineConfig) (core.PipelineUUID, error)
	RunPipeline(pID core.PipelineUUID) error
	EventLoop(ctx context.Context) error
}

// etlManager ... ETL manager
type etlManager struct {
	ctx context.Context

	analyzer Analyzer
	dag      ComponentGraph
	store    EtlStore

	engineChan chan core.InvariantInput

	registry registry.Registry
	wg       sync.WaitGroup
}

// NewManager ... Initializer
func NewManager(ctx context.Context, analyzer Analyzer, cRegistry registry.Registry, ec chan core.InvariantInput) (Manager, func()) {
	dag := NewComponentGraph()

	m := &etlManager{
		analyzer:   analyzer,
		ctx:        ctx,
		dag:        dag,
		store:      newEtlStore(),
		registry:   cRegistry,
		engineChan: ec,
		wg:         sync.WaitGroup{},
	}

	shutDown := func() { // Iterate and kill active pipelines
		for _, pl := range m.store.GetAllPipelines() {
			logging.WithContext(ctx).
				Info("Shuting down pipeline",
					zap.String(core.PUUIDKey, pl.UUID().String()))

			if err := pl.Close(); err != nil {
				logging.WithContext(ctx).
					Error("Failed to close pipeline",
						zap.String(core.PUUIDKey, pl.UUID().String()))
			}
		}
		logging.WithContext(ctx).Debug("Waiting for all component routines to end")
		m.wg.Wait()

		logging.WithContext(ctx).Debug("Closing component event channel")
	}

	return m, shutDown
}

// GetRegister ... Returns a data register for a given register type
func (em *etlManager) GetRegister(rt core.RegisterType) (*core.DataRegister, error) {
	return em.registry.GetRegister(rt)
}

// CreateDataPipeline ... Creates an ETL data pipeline provided a pipeline configuration
func (em *etlManager) CreateDataPipeline(cfg *core.PipelineConfig) (core.PipelineUUID, error) {
	// NOTE - If some of these early sub-system operations succeed but lower function
	// code logic fails, then some rollback will need be triggered to undo prior applied state operations
	logger := logging.WithContext(em.ctx)

	depPath, err := em.registry.GetDependencyPath(cfg.DataType)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	pUUID := depPath.GeneratePipelineUUID(cfg.PipelineType, cfg.Network)

	components, err := em.getComponents(cfg, depPath)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	logger.Debug("constructing pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	pipeline, err := NewPipeline(cfg, pUUID, components)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	pipelines := em.store.GetExistingPipelinesByPID(pUUID.PID)

	for _, pl := range pipelines {
		p, err := em.store.GetPipelineFromPUUID(pl)
		if err != nil {
			return core.NilPipelineUUID(), err
		}

		if em.analyzer.Mergable(pipeline, p) { // Deploy invariants to existing pipelines instead
			// This is a bit hacky since we aren't actually merging the pipelines
			return p.UUID(), nil
		}
	}

	// Bind communication route between pipeline and risk engine
	if err := pipeline.AddEngineRelay(em.engineChan); err != nil {
		return core.NilPipelineUUID(), err
	}

	if err := em.dag.AddComponents(pipeline.Components()); err != nil {
		return core.NilPipelineUUID(), err
	}

	em.store.AddPipeline(pUUID, pipeline)

	if len(components) == 1 {
		return pUUID, nil
	}

	for i := 1; i < len(components); i++ {
		em.dag.AddEdge(components[i].UUID(), components[i-1].UUID())
	}

	return pUUID, nil
}

// RunPipeline ... Runs pipeline session for some provided pUUID
func (em *etlManager) RunPipeline(pUUID core.PipelineUUID) error {
	pipeline, err := em.store.GetPipelineFromPUUID(pUUID)
	if err != nil {
		return err
	}

	logging.WithContext(em.ctx).Info("Running pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	return pipeline.RunPipeline(&em.wg)
}

// EventLoop ... Driver ran as separate go routine
func (em *etlManager) EventLoop(ctx context.Context) error {
	logger := logging.WithContext(ctx)

	for {
		<-ctx.Done()
		logger.Info("Receieved shutdown request")
		return nil
	}
}

// getComponents ... Returns all components provided a slice of register definitions
func (em *etlManager) getComponents(cfg *core.PipelineConfig,
	depPath core.RegisterDependencyPath) ([]component.Component, error) {
	components := make([]component.Component, 0)
	// prevUUID := core.NilComponentUUID()

	for _, register := range depPath.Path {
		cUUID := core.MakeComponentUUID(cfg.PipelineType, register.ComponentType, register.DataType, cfg.Network)

		c, err := inferComponent(em.ctx, cfg, cUUID, register)
		if err != nil {
			return []component.Component{}, err
		}

		components = append(components, c)
	}

	return components, nil
}

// inferComponent ... Constructs a component provided a data register definition
func inferComponent(ctx context.Context, cfg *core.PipelineConfig, cUUID core.ComponentUUID,
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

		return init(ctx, cfg.ClientConfig, opts...)

	case core.Pipe:
		init, success := register.ComponentConstructor.(component.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Pipe.String()))
		}

		return init(ctx, cfg.ClientConfig, opts...)

	case core.Aggregator:
		return nil, fmt.Errorf("aggregator component has yet to be implemented")

	default:
		return nil, fmt.Errorf(unknownCompType, register.ComponentType.String())
	}
}

func (em *etlManager) mergePipelines(p1, p2 Pipeline) {
}
