package pipeline

import (
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

type Pipeline interface {
	UUID() core.PipelineUUID
	Close() error
	Components() []component.Component
	RunPipeline(wg *sync.WaitGroup) error

	AddEngineRelay(engineChan chan core.InvariantInput) error
}

type Option = func(*pipeline)

type pipeline struct {
	uuid core.PipelineUUID

	aState ActivityState
	pType  core.PipelineType //nolint:unused // will be implemented soon

	components []component.Component
}

// NewPipeLine ... Initializer
func NewPipeLine(pUUID core.PipelineUUID, comps []component.Component, opts ...Option) (Pipeline, error) {
	pl := &pipeline{
		uuid:       pUUID,
		components: comps,
		aState:     Booting,
	}

	for _, opt := range opts {
		opt(pl)
	}

	return pl, nil
}

// Components ... Returns slice of all constituent components
func (pl *pipeline) Components() []component.Component {
	return pl.components
}

// UUID ... Returns pipeline UUID
func (pl *pipeline) UUID() core.PipelineUUID {
	return pl.uuid
}

// AddEngineRelay ... Adds a relay to the pipeline that forces it to send transformed invariant input
// to a risk engine
func (pl *pipeline) AddEngineRelay(engineChan chan core.InvariantInput) error {
	lastComponent := pl.components[len(pl.components)-1]
	eir := core.NewEngineRelay(pl.uuid, engineChan)

	return lastComponent.AddRelay(eir)
}

// RunPipeline  ... Spawns and manages component event loops
// for some pipeline
func (pl *pipeline) RunPipeline(wg *sync.WaitGroup) error {
	for _, comp := range pl.components {
		wg.Add(1)

		go func(c component.Component, wg *sync.WaitGroup) {
			defer wg.Done()

			logging.NoContext().
				Debug("Attempting to start component event loop",
					zap.String(core.CUUIDKey, c.UUID().String()),
					zap.String(core.PUUIDKey, pl.uuid.String()))

			if err := c.EventLoop(); err != nil {
				logging.NoContext().Error("Obtained error from event loop", zap.Error(err),
					zap.String(core.CUUIDKey, c.UUID().String()),
					zap.String(core.PUUIDKey, pl.uuid.String()))
			}
		}(comp, wg)
	}

	return nil
}

// Close ...
func (pl *pipeline) Close() error {
	for _, comp := range pl.components {
		if comp.ActivityState() != component.Terminated {
			logging.NoContext().
				Debug("Shutting down pipeline component",
					zap.String(core.CUUIDKey, comp.UUID().String()),
					zap.String(core.PUUIDKey, pl.uuid.String()))

			if err := comp.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
