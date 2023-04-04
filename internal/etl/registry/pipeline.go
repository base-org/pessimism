package registry

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/models"
)

type RegisterPipeline struct {
	ctx        context.Context
	components []pipeline.Component
	dataType   models.RegisterType
}

func NewRegisterPipeline(ctx context.Context, cfg *config.RegisterPipelineConfig) (*RegisterPipeline, error) {
	log.Printf("Constructing register pipeline for %s", cfg.DataType)

	register, err := GetRegister(cfg.DataType)
	if err != nil {
		return nil, err
	}

	components := make([]pipeline.Component, 0)

	for _, register := range append([]*DataRegister{register}, register.Dependencies...) {

		component, err := inferComponent(ctx, cfg, register)
		if err != nil {
			return nil, err
		}

		components = append(components, component)
	}

	return &RegisterPipeline{
		ctx:        ctx,
		components: components,
		dataType:   cfg.DataType,
	}, nil
}

func (rp *RegisterPipeline) RunPipeline(wg *sync.WaitGroup) error {

	for index, component := range rp.components {
		if component.Type() == models.Pipe {
			// Assumptions
			// 1 - Pipe component node always has a previous node in the DAG sequence
			// If not an index out of bounds exception would occur

			entryChan := component.EntryPoints()[0]

			err := rp.components[index+1].AddDirective(component.ID(), entryChan)
			if err != nil {

			}

		}

	}

	for _, component := range rp.components {
		wg.Add(1)

		go func(c pipeline.Component, wg *sync.WaitGroup) {
			log.Printf("Starting event loop for component: %s", c)

			defer wg.Done()

			if err := c.EventLoop(); err != nil {
				log.Printf("Got error from event loop: %s", err.Error())
			}

		}(component, wg)
	}

	return nil

}

func inferComponent(ctx context.Context, cfg *config.RegisterPipelineConfig, register *DataRegister) (pipeline.Component, error) {
	log.Printf("Constructing %s component for register %s", register.ComponentType, register.DataType)

	switch register.ComponentType {
	case models.Oracle:
		init, success := register.ComponentConstructor.(pipeline.OracleConstructorFunc)
		if !success {
			return nil, fmt.Errorf("Could not cast constructor to oracle constructor type")
		}

		// NOTE ... We assume at most 1 oracle per register pipeline
		return init(ctx, cfg.PipelineType, cfg.OracleCfg)

	case models.Pipe:
		init, success := register.ComponentConstructor.(pipeline.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf("Could not cast constructor to pipe constructor type")
		}

		return init(ctx)

	case models.Conveyor:
		return nil, fmt.Errorf("Conveyor component has yet to be implemented")

	default:
		return nil, fmt.Errorf("Unknown component type provided")
	}
}

func (rp *RegisterPipeline) AddDirective(componentID models.ComponentID, outChan chan models.TransitData) error {
	lastComponent := rp.components[0]
	log.Printf("Adding directive for component %s %s", lastComponent.ID(), lastComponent.Type())
	return lastComponent.AddDirective(componentID, outChan)
}
