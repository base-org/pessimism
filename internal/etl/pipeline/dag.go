package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/models"
)

type componentEntry struct {
	comp      component.Component
	outType   models.RegisterType
	pipeLines []models.ID
}

type DAG struct {
	componentMap map[models.ID]componentEntry
	edgeMap      map[models.ID][]models.ID

	pipeLines map[models.ID]*pipeLine

	sync.WaitGroup
}

func NewDAG() *DAG {
	return &DAG{
		componentMap: make(map[models.ID]componentEntry),
		edgeMap:      make(map[models.ID][]models.ID),
		pipeLines:    make(map[models.ID]*pipeLine),
	}
}

// AddEdge ... Adds edge between two already constructed component nodes
// C1 --> C2
// C1.Router --channel--> C2.entry_point where C1.output_type == C2.entry_point
func (dag *DAG) AddEdge(cID1, cID2 models.ID) error {
	entry1, found := dag.componentMap[cID1]
	if !found {
		return fmt.Errorf("Could not find a valid component in mapping")
	}

	entry2, found := dag.componentMap[cID2]
	if !found {
		return fmt.Errorf("Could not find a valid component in mapping")
	}

	// Instructs component_1
	log.Printf("Getting entry for output type %s", entry1.outType)

	if err := entry2.comp.CreateEntryPoint(entry1.outType); err != nil {
		return err
	}

	entryChan, err := entry2.comp.GetEntryPoint(entry1.outType)
	if err != nil {
		return err
	}

	if err := entry1.comp.AddDirective(cID2, entryChan); err != nil {
		return err
	}

	// Update edge mapping with new link
	dag.edgeMap[cID1] = append(dag.edgeMap[cID1], cID2)

	return nil
}

func (dag *DAG) RemoveEdge(cID1, cID2 models.ID) error {
	// TODO
	return nil
}

func (dag *DAG) CreateRegisterPipeline(ctx context.Context, cfg *config.RegisterPipelineConfig) (models.ID, error) {
	log.Printf("Constructing register pipeline for %s", cfg.DataType)

	register, err := registry.GetRegister(cfg.DataType)
	if err != nil {
		return models.NilID(), err
	}

	components := make([]component.Component, 0)
	registers := append([]*registry.DataRegister{register}, register.Dependencies...)
	log.Printf("%+v", registers)

	var prevID models.ID = models.NilID()

	for i, register := range registers {
		// NOTE - This doesn't consider the circumstance where a requested pipeline already exists but requires some backfill to run
		id := models.Strings2ID(cfg.PipelineType, string(cfg.DataType))
		if err != nil {
			return models.NilID(), err
		}

		var component component.Component

		if _, exists := dag.edgeMap[id]; !exists {
			// CASE 1 : Pipeline component doesn't exists within DAG

			component, err = inferComponent(ctx, cfg, id, register)
			if err != nil {
				return models.NilID(), err
			}

			dag.edgeMap[component.ID()] = make([]models.ID, 0)
			dag.componentMap[component.ID()] = componentEntry{
				outType:   register.DataType,
				comp:      component,
				pipeLines: []models.ID{id},
			}

		} else {
			component = dag.componentMap[id].comp

		}

		if i != 0 { // IE we've passed the pipeline's origin node
			if err := dag.AddEdge(id, prevID); err != nil {
				return models.NilID(), err
			}
		}

		prevID = component.ID()
		components = append(components, component)
	}

	pID, err := generatePipelineID(components...)
	if err != nil {
		return models.NilID(), err
	}

	pipeLine, err := newPipeLine(pID, components...)
	if err != nil {
		return models.NilID(), err
	}

	dag.pipeLines[pID] = pipeLine
	// TODO - Update pipeline entries with component entry struct within componentMap

	return pipeLine.id, nil
}

func (dag *DAG) RunPipeline(id models.ID) error {
	pipeLine, found := dag.pipeLines[id]
	if !found {
		return fmt.Errorf("Could not find pipeline for id: %s", id)
	}

	return pipeLine.RunPipeline(&dag.WaitGroup)
}

func (dag *DAG) AddPipelineDirective(pID models.ID, outID models.ID, outChan chan models.TransitData) error {
	pipeLine, found := dag.pipeLines[pID]
	if !found {
		return fmt.Errorf("Could not find pipeline for id: %s", pID)
	}

	return pipeLine.components[len(pipeLine.components)-1].AddDirective(outID, outChan)
}

// inferComponent ...
func inferComponent(ctx context.Context, cfg *config.RegisterPipelineConfig, id models.ID,
	register *registry.DataRegister) (component.Component, error) {
	log.Printf("Constructing %s component for register %s", register.ComponentType, register.DataType)

	switch register.ComponentType {
	case models.Oracle:
		init, success := register.ComponentConstructor.(component.OracleConstructorFunc)
		if !success {
			return nil, fmt.Errorf("could not cast constructor to oracle constructor type")
		}

		// NOTE ... We assume at most 1 oracle per register pipeline
		return init(ctx, cfg.PipelineType, cfg.OracleCfg, component.WithID(id))

	case models.Pipe:
		init, success := register.ComponentConstructor.(component.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf("could not cast constructor to pipe constructor type")
		}

		return init(ctx)

	case models.Aggregator:
		return nil, fmt.Errorf("aggregator component has yet to be implemented")

	default:
		return nil, fmt.Errorf("unknown component type provided")
	}
}
