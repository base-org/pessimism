package pipeline

import (
	"context"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

/*
 Edge cases:
 1 - Synchronization required for a pipeline from some starting block height
 2 - Live config is passed through to a register pipeline config that requires a backfill,
 	resulting in component collision within the DAG.
 3 -
*/

type PipeLine interface {
	// EventLoop ... Pipeline driver function; spun up as separate go routine
	ID() core.PipelineID
	EventLoop() error
	RunPipeline(wg *sync.WaitGroup) error
	AddDirective(cID core.ComponentID, outChan chan core.TransitData) error
}

type Option = func(*pipeLine)

type pipeLine struct {
	ctx context.Context
	id  core.PipelineID

	aState ActivityState
	pType  core.PipelineType //nolint:unused // will be implemented soon

	components []component.Component
}

func NewPipeLine(id core.PipelineID, comps []component.Component, opts ...Option) (PipeLine, error) {
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

func (pl *pipeLine) ID() core.PipelineID {
	return pl.id
}

func (pl *pipeLine) AddDirective(cID core.ComponentID, outChan chan core.TransitData) error {
	comp := pl.components[0]
	log.Printf("Adding output directive between components (%s) --> (%s)", comp.ID().String(), cID.String())

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

func (pl *pipeLine) String() string {
	str := ""

	for i := len(pl.components) - 1; i >= 0; i-- {
		comp := pl.components[i]
		str += "(" + comp.ID().String() + ")"
		if i != 0 {
			str += "->"
		}
	}

	return str
}

func (pl *pipeLine) EventLoop() error {
	// TODO - Add component sampling logic
	// I.E, Components in a pipeline should be checked for activity state changes
	// Critical for understanding when things like "syncing" or backfilling have completed for some
	// live invariant pipeline

	for { //nolint:gosimple // will soon be extended to other go channels
		select {
		case <-pl.ctx.Done():
			return nil
		}
	}
}
