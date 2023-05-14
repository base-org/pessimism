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

// Manager ...
type Manager interface {
	CreateDataPipeline(cfg *core.PipelineConfig) (core.PipelineUUID, error)
	RunPipeline(pID core.PipelineUUID) error
	EventLoop(ctx context.Context) error
}

// etlManager ... Holds
type etlManager struct {
	ctx context.Context

	dag      ComponentGraph
	etlStore EtlStore

	engineChan    chan core.InvariantInput
	compEventChan chan component.StateChange

	wg sync.WaitGroup
}

// NewManager ... Initializer
func NewManager(ctx context.Context, ec chan core.InvariantInput) (Manager, func()) {
	dag := NewComponentGraph()

	m := &etlManager{
		ctx:           ctx,
		dag:           dag,
		etlStore:      newEtlStore(),
		engineChan:    ec,
		compEventChan: make(chan component.StateChange),
		wg:            sync.WaitGroup{},
	}

	shutDown := func() { // Iterate and kill active pipelines
		for _, pl := range m.etlStore.GetAllPipelines() {
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
		close(m.compEventChan)
	}

	return m, shutDown
}

// CreateDataPipeline ... Creates an ETL data pipeline provided a pipeline configuration
func (em *etlManager) CreateDataPipeline(cfg *core.PipelineConfig) (core.PipelineUUID, error) {
	// NOTE - If some of these early sub-system operations succeed but lower function
	// code logic fails, then some rollback will need be triggered to undo prior applied state operations
	logger := logging.WithContext(em.ctx)

	register, err := registry.GetRegister(cfg.DataType)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	depPath := register.GetDependencyPath()
	pUUID := depPath.GeneratePipelineUUID(cfg.PipelineType, cfg.Network)

	components, err := em.getComponents(cfg, depPath)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	logger.Debug("constructing pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	pipeline, err := NewPipeLine(pUUID, components)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	if err := pipeline.AddEngineRelay(em.engineChan); err != nil {
		return core.NilPipelineUUID(), err
	}

	if err := em.dag.AddComponents(pipeline.Components()); err != nil {
		return core.NilPipelineUUID(), err
	}

	em.etlStore.AddPipeline(pUUID, pipeline)

	return pUUID, nil
}

// RunPipeline ... Runs pipeline session for some provided pUUID
func (em *etlManager) RunPipeline(pUUID core.PipelineUUID) error {
	pipeline, err := em.etlStore.GetPipelineFromPUUID(pUUID)
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
		select {
		case <-ctx.Done():
			logger.Info("Receieved shutdown request")
			return nil

		case stateChange := <-em.compEventChan:
			// TODO(#35): No ETL Management Procedure Exists
			// for Handling Component State Changes

			logger.Info("Received component state change request",
				zap.String("from", stateChange.From.String()),
				zap.String("to", stateChange.To.String()),
				zap.String(core.CUUIDKey, stateChange.ID.String()))

			_, err := em.etlStore.GetPipelineUUIDs(stateChange.ID)
			if err != nil {
				logger.Error("Could not fetch pipeline IDs for comp state change")
			}
		}
	}
}

// getComponents ... Returns all components provided a slice of register definitions
func (em *etlManager) getComponents(cfg *core.PipelineConfig,
	depPath core.RegisterDependencyPath) ([]component.Component, error) {
	components := make([]component.Component, 0)
	prevID := core.NilComponentUUID()

	for _, register := range depPath.Path {
		// TODO(#30): Pipeline Collisions Occur When They Shouldn't
		cUUID := core.MakeComponentUUID(cfg.PipelineType, register.ComponentType, register.DataType, cfg.Network)

		c, err := inferComponent(em.ctx, cfg, cUUID, register)
		if err != nil {
			return []component.Component{}, err
		}

		if prevID != core.NilComponentUUID() { // IE we've passed the pipeline's last path node; start adding edges (n, n-1)
			if err := em.dag.AddEdge(cUUID, prevID); err != nil {
				return []component.Component{}, err
			}
		}

		prevID = c.UUID()
		components = append(components, c)
	}

	return components, nil
}

// inferComponent ... Constructs a component provided a data register definition
func inferComponent(ctx context.Context, cfg *core.PipelineConfig, id core.ComponentUUID,
	register *core.DataRegister) (component.Component, error) {
	logging.WithContext(ctx).Debug("constructing component",
		zap.String("type", register.ComponentType.String()),
		zap.String("outdata_type", register.DataType.String()))

	switch register.ComponentType {
	case core.Oracle:
		init, success := register.ComponentConstructor.(component.OracleConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Oracle.String()))
		}

		return init(ctx, cfg.PipelineType, cfg.OracleCfg,
			component.WithID(id))

	case core.Pipe:
		init, success := register.ComponentConstructor.(component.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Pipe.String()))
		}

		return init(ctx, component.WithID(id))

	case core.Aggregator:
		return nil, fmt.Errorf("aggregator component has yet to be implemented")

	default:
		return nil, fmt.Errorf(unknownCompType, register.ComponentType.String())
	}
}
