package pipeline

import (
	"fmt"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

type PipeLine interface {
	ID() core.PipelineUUID
	Components() []component.Component
	// EventLoop ... Pipeline driver function; spun up as separate go routine
	RunPipeline(wg *sync.WaitGroup) error
	UpdateState(as ActivityState) error

	AddDirective(cID core.ComponentUUID, outChan chan core.TransitData) error
}

type Option = func(*pipeLine)

type pipeLine struct {
	id core.PipelineUUID

	aState ActivityState
	pType  core.PipelineType //nolint:unused // will be implemented soon

	components []component.Component
}

func NewPipeLine(id core.PipelineUUID, comps []component.Component, opts ...Option) (PipeLine, error) {
	pl := &pipeLine{
		id:         id,
		components: comps,
		aState:     Booting,
	}

	for _, opt := range opts {
		opt(pl)
	}

	return pl, nil
}

func (pl *pipeLine) Components() []component.Component {
	return pl.components
}

func (pl *pipeLine) ID() core.PipelineUUID {
	return pl.id
}

func (pl *pipeLine) AddDirective(cID core.ComponentUUID, outChan chan core.TransitData) error {
	comp := pl.components[0]

	return comp.AddEgress(cID, outChan)
}

func (pl *pipeLine) RunPipeline(wg *sync.WaitGroup) error {
	for _, comp := range pl.components {
		wg.Add(1)

		go func(c component.Component, wg *sync.WaitGroup) {
			log.Printf("Attempting to run component (%s) with activity state = %s", c.ID().String(), c.ActivityState())
			if c.ActivityState() != component.Inactive {
				return
			}

			defer wg.Done()

			if err := c.EventLoop(); err != nil {
				log.Printf("Got error from event loop: %s", err.Error())
			}
		}(comp, wg)
	}

	return nil
}

// TerminatePipeline ...
func (pl *pipeLine) Terminate(_ *sync.WaitGroup) error {
	return nil
}

func (pl *pipeLine) UpdateState(as ActivityState) error {
	if as == pl.aState {
		return fmt.Errorf("state is already set")
	}

	pl.aState = as
	return nil
}
