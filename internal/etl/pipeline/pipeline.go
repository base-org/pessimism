package pipeline

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/models"
)

/*
 Edge cases:
 1 - Synchronization required for a pipeline from some starting block height
 2 - Live config is passed through to a register pipeline config that requires a backfill

*/

type PipeLineState = string

const (
	Booting PipeLineState = "booting"
	Syncing PipeLineState = "syncing"
	Active  PipeLineState = "active"
	Crashed PipeLineState = "crashed"
)

type pipeLine struct {
	ctx context.Context
	id  models.ID

	pState PipeLineState
	pType  models.PipelineType

	components []component.Component
}

func generatePipelineID(comps ...component.Component) (models.ID, error) {
	ids := make([]string, len(comps))

	for i, comp := range comps {
		ids[i] = fmt.Sprintf("%+v", comp.ID())
	}

	id := models.Strings2ID(ids...)

	return id, nil
}

func newPipeLine(id models.ID, comps ...component.Component) (*pipeLine, error) {

	return &pipeLine{
		id:         id,
		components: comps,
		pState:     Booting,
	}, nil
}

func (pl *pipeLine) RunPipeline(wg *sync.WaitGroup) error {
	for _, comp := range pl.components {
		if comp.GetActivityState() != component.Inactive {
			continue
		}

		wg.Add(1)

		go func(c component.Component, wg *sync.WaitGroup) {
			log.Printf("Starting event loop for component: %s", c.ID())

			defer wg.Done()

			if err := c.EventLoop(); err != nil {
				log.Printf("Got error from event loop: %s", err.Error())
			}
		}(comp, wg)
	}

	return nil
}
