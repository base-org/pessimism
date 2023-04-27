package pipeline

import (
	"fmt"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

type Pipeline interface {
	ID() core.PipelineUUID
	Components() []component.Component
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

// NewPipeLine ... Initializer
func NewPipeLine(id core.PipelineUUID, comps []component.Component, opts ...Option) (Pipeline, error) {
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

// Components ... Returns slice of all constituent components
func (pl *pipeLine) Components() []component.Component {
	return pl.components
}

// ID ... Returns pipeline ID
func (pl *pipeLine) ID() core.PipelineUUID {
	return pl.id
}

// NOTE: This function should be deleted once risk engine is introduced
// AddDirective ... Adds an egress from the final component of a pipeline
// path to an arbitrary channel
func (pl *pipeLine) AddDirective(cID core.ComponentUUID, outChan chan core.TransitData) error {
	comp := pl.components[0]

	return comp.AddEgress(cID, outChan)
}

// RunPipeline  ... Spawns and manages component event loops
// for some pipeline
func (pl *pipeLine) RunPipeline(wg *sync.WaitGroup) error {
	for _, comp := range pl.components {
		wg.Add(1)

		go func(c component.Component, wg *sync.WaitGroup) {
			defer wg.Done()

			log.Printf("Attempting to run component (%s) with activity state = %s", c.ID().String(), c.ActivityState())
			if c.ActivityState() != component.Inactive { // Component already active
				return
			}

			if err := c.EventLoop(); err != nil {
				log.Printf("Got error from event loop: %s", err.Error())
			}
		}(comp, wg)
	}

	return nil
}

// Terminate ...
func (pl *pipeLine) Terminate(_ *sync.WaitGroup) error {
	// TODO: implement
	return nil
}

func (pl *pipeLine) UpdateState(as ActivityState) error {
	if as == pl.aState {
		return fmt.Errorf("state is already set")
	}

	pl.aState = as
	return nil
}
