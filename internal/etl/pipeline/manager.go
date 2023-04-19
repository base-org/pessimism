package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/config"
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

	compEventChan chan component.StateChange

	wg *sync.WaitGroup
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:           ctx,
		dag:           newGraph(),
		pRegistry:     newPipeRegistry(),
		compEventChan: make(chan component.StateChange),
		wg:            &sync.WaitGroup{},
	}
}

func (manager *Manager) CreatePipeline(ctx context.Context,
	cfg *config.PipelineConfig) (core.PipelineID, error) {
	logger := logging.WithContext(manager.ctx)

	register, err := registry.GetRegister(cfg.DataType)
	if err != nil {
		return core.NilPipelineID(), err
	}

	components := make([]component.Component, 0)
	registers := append([]*core.DataRegister{register}, register.Dependencies...)

	prevID := core.NilCompID()
	lastReg := registers[len(registers)-1]

	cID1 := core.MakeComponentID(cfg.PipelineType, registers[0].ComponentType, registers[0].DataType, cfg.Network)
	cID2 := core.MakeComponentID(cfg.PipelineType, lastReg.ComponentType, lastReg.DataType, cfg.Network)
	pID := core.MakePipelineID(cfg.PipelineType, cID1, cID2)

	logger.Debug("constructing pipeline",
		zap.String("ID", pID.String()))

	for i, register := range registers {
		// NOTE - This doesn't consider the circumstance where
		// a requested pipeline already exists but requires some backfill to run
		// TODO(#30): Pipeline Collisions Occur When They Shouldn't
		cID := core.MakeComponentID(cfg.PipelineType, register.ComponentType, register.DataType, cfg.Network)
		if err != nil {
			return core.NilPipelineID(), err
		}

		if !manager.dag.componentExists(cID) {
			comp, err := inferComponent(ctx, cfg, cID, register, manager.compEventChan)
			if err != nil {
				return core.NilPipelineID(), err
			}
			if err = manager.dag.addComponent(cID, comp); err != nil {
				return core.NilPipelineID(), err
			}
		}

		component, err := manager.dag.getComponent(cID)
		if err != nil {
			return core.NilPipelineID(), err
		}

		if i != 0 { // IE we've passed the pipeline's last path node; start adding edges
			if err := manager.dag.addEdge(cID, prevID); err != nil {
				return core.NilPipelineID(), err
			}
		}

		prevID = component.ID()
		components = append(components, component)
	}

	pipeLine, err := NewPipeLine(pID, components)
	if err != nil {
		return core.NilPipelineID(), err
	}

	manager.pRegistry.addPipeline(pID, pipeLine, cfg.PipelineType)

	return pID, nil
}

func (manager *Manager) RunPipeline(pID core.PipelineID) error {
	pipeLine, err := manager.pRegistry.getPipeline(pID)
	if err != nil {
		return err
	}

	log.Printf("[%s] Running pipeline", pipeLine.ID().String())
	return pipeLine.RunPipeline(manager.wg)
}

func (manager *Manager) AddPipelineDirective(pID core.PipelineID,
	cID core.ComponentID, outChan chan core.TransitData) error {
	pipeLine, err := manager.pRegistry.getPipeline(pID)
	if err != nil {
		return err
	}
	return pipeLine.AddDirective(cID, outChan)
}

func (m *Manager) EventLoop(ctx context.Context) {
	logger := logging.WithContext(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Received shutdown", zap.String("Abstraction", "etl-manager"))

		case stateChange := <-m.compEventChan:
			_, err := m.pRegistry.fetchCompPipelineIDs(stateChange.ID)
			if err != nil {
				logger.Error("Could not fetch pipeline IDs for comp state change")
			}

			logger.Info("Received component state change requeset")
		}

	}

}

func inferComponent(ctx context.Context, cfg *config.PipelineConfig, id core.ComponentID,
	register *core.DataRegister, eventCh chan component.StateChange) (component.Component, error) {
	log.Printf("constructing %s component for register %s", register.ComponentType, register.DataType)

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
