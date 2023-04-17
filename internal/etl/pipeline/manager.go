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
	pipeLines map[core.PipelineID]PipeLine
	wg        *sync.WaitGroup
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx:       ctx,
		dag:       newGraph(),
		pipeLines: make(map[core.PipelineID]PipeLine, 0),
		wg:        &sync.WaitGroup{},
	}
}

func (manager *Manager) CreateRegisterPipeline(ctx context.Context,
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

	logger.Debug("constructing register pipeline",
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
			comp, err := inferComponent(ctx, cfg, cID, register)
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

	manager.pipeLines[pID] = pipeLine

	return pID, nil
}

func (manager *Manager) RunPipeline(id core.PipelineID) error {
	pipeLine, found := manager.pipeLines[id]
	if !found {
		return fmt.Errorf("could not find pipeline for id: %s", id)
	}

	log.Printf("[%s] Running pipeline", pipeLine.ID().String())
	return pipeLine.RunPipeline(manager.wg)
}

func (manager *Manager) AddPipelineDirective(pID core.PipelineID,
	cID core.ComponentID, outChan chan core.TransitData) error {
	pipeLine, found := manager.pipeLines[pID]
	if !found {
		return fmt.Errorf("could not find pipeline for id: %s", pID)
	}

	return pipeLine.AddDirective(cID, outChan)
}

func inferComponent(ctx context.Context, cfg *config.PipelineConfig, id core.ComponentID,
	register *core.DataRegister) (component.Component, error) {
	log.Printf("constructing %s component for register %s", register.ComponentType, register.DataType)

	switch register.ComponentType {
	case core.Oracle:
		init, success := register.ComponentConstructor.(component.OracleConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Oracle.String()))
		}

		// NOTE ... We assume at most 1 oracle per register pipeline
		return init(ctx, cfg.PipelineType, cfg.OracleCfg, component.WithID(id))

	case core.Pipe:
		init, success := register.ComponentConstructor.(component.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf(fmt.Sprintf(couldNotCastErr, core.Pipe.String()))
		}

		return init(ctx, component.WithID(id))

	case core.Aggregator:
		return nil, fmt.Errorf("aggregator component has yet to be implemented")

	default:
		return nil, fmt.Errorf("unknown component type provided")
	}
}
