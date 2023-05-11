package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

type Manager struct {
	ctx context.Context

	dag       *cGraph
	pRegistry *pipeRegistry

	closeChan     chan int
	engineChan    chan core.InvariantInput
	compEventChan chan component.StateChange

	wg *sync.WaitGroup
}

// NewManager ... Initializer
func NewManager(ctx context.Context, ec chan core.InvariantInput, wg *sync.WaitGroup) (*Manager, func()) {
	dag := newGraph()

	m := &Manager{
		ctx:           ctx,
		dag:           dag,
		closeChan:     make(chan int, 1),
		pRegistry:     newPipeRegistry(),
		engineChan:    ec,
		compEventChan: make(chan component.StateChange),
		wg:            wg,
	}

	shutDown := func() {
		for id := range dag.edges() {
			c, err := dag.getComponent(id)
			if err != nil {
				logging.NoContext().Error("Could not get component during shutdown", zap.Error(err))
			}

			if err := c.Close(); err != nil {
				logging.NoContext().Error("Could not close component", zap.Error(err))
			}
		}

		m.closeChan <- 0
	}

	return m, shutDown
}

// CreateDataPipeline ... Creates an ETL data pipeline provided a pipeline configuration
func (m *Manager) CreateDataPipeline(cfg *core.PipelineConfig) (core.PipelineUUID, error) {
	// NOTE - If some of these early sub-system operations succeed but lower function
	// code logic fails, then some rollback will need be triggered to undo prior applied state operations
	logger := logging.WithContext(m.ctx)

	outputReg, err := registry.GetRegister(cfg.DataType)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	depPath := outputReg.GetDependencyPath()
	pUUID := depPath.GeneratePipelineUUID(cfg.PipelineType, cfg.Network)

	components, err := m.getComponents(cfg, depPath)
	if err != nil {
		return core.NilPipelineUUID(), nil
	}

	logger.Debug("constructing pipeline",
		zap.String("ID", pUUID.String()))

	pipeLine, err := NewPipeLine(pUUID)
	if err != nil {
		return core.NilPipelineUUID(), err
	}

	comps := pipeLine.Components()
	lastComp := comps[len(comps)-1]

	relay := core.NewEngineRelay(pUUID, m.engineChan)

	// Route pipeline output to risk engine as invariant input
	err = lastComp.AddRelay(relay)
	if err != nil {
		return core.NilPipelineUUID(), err
	}
	m.pRegistry.addPipeline(pUUID, pipeLine)

	return pUUID, nil
}

func (m *Manager) RunPipeline(pID core.PipelineUUID) error {
	pipeLine, err := m.pRegistry.getPipeline(pID)
	if err != nil {
		return err
	}

	log.Printf("[%s] Running pipeline", pipeLine.ID().String())
	return pipeLine.RunPipeline(m.wg)
}

func (m *Manager) EventLoop(ctx context.Context) {
	logger := logging.WithContext(ctx)

	for {
		select {
		case <-m.closeChan:
			logger.Info("etlManager receieved shutdown request")
			return

		case stateChange := <-m.compEventChan:
			// TODO(#35): No ETL Management Procedure Exists
			// for Handling Component State Changes

			_, err := m.pRegistry.getPipelineUUIDs(stateChange.ID)
			if err != nil {
				logger.Error("Could not fetch pipeline IDs for comp state change")
			}

			logger.Info("Received component state change request")
		}
	}
}

// getComponents ... Returns all components provided a slice of register definitions
func (m *Manager) getComponents(cfg *core.PipelineConfig,
	depPath core.RegisterDependencyPath) ([]component.Component, error) {
	components := make([]component.Component, 0)
	prevID := core.NilComponentUUID()

	for i, register := range registers {
		// NOTE - This doesn't consider the circumstance where
		// a requested pipeline already exists but requires some backfill to run
		// TODO(#30): Pipeline Collisions Occur When They Shouldn't
		cID := core.MakeComponentUUID(cfg.PipelineType, register.ComponentType, register.DataType, cfg.Network)

		if !m.dag.componentExists(cID) {
			comp, err := inferComponent(m.ctx, cfg, cID, register, m.compEventChan)
			if err != nil {
				return []component.Component{}, err
			}
			if err = m.dag.addComponent(cID, comp); err != nil {
				return []component.Component{}, err
			}
		}

		c, err := m.dag.getComponent(cID)
		if err != nil {
			return []component.Component{}, err
		}

		if i != 0 { // IE we've passed the pipeline's last path node; start adding edges
			if err := m.dag.addEdge(cID, prevID); err != nil {
				return []component.Component{}, err
			}
		}

		prevID = c.ID()
		components = append(components, c)
	}

	return components, nil
}

// inferComponent ... Constructs a component provided a data register definition
func inferComponent(ctx context.Context, cfg *core.PipelineConfig, id core.ComponentUUID,
	register *core.DataRegister, eventCh chan component.StateChange) (component.Component, error) {
	logging.WithContext(ctx).Debug("constructing component",
		zap.String("type", register.ComponentType.String()),
		zap.String("outdata_type", register.DataType.String()))

	switch register.ComponentType {
	case core.Oracle:
		init, success := register.ComponentConstructor.(component.OracleConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Oracle.String()))
		}

		// NOTE ... We assume at most 1 oracle per register pipeline
		return init(ctx, cfg.PipelineType, cfg.OracleCfg,
			component.WithID(id), component.WithEventChan(eventCh))

	case core.Pipe:
		init, success := register.ComponentConstructor.(component.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Pipe.String()))
		}

		return init(ctx, component.WithID(id), component.WithEventChan(eventCh))

	case core.Aggregator:
		return nil, fmt.Errorf("aggregator component has yet to be implemented")

	default:
		return nil, fmt.Errorf("unknown component type provided")
	}
}
