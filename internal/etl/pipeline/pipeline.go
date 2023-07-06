package pipeline

import (
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Pipeline ... Pipeline interface
type Pipeline interface {
	Config() *core.PipelineConfig
	UUID() core.PUUID
	Close() error
	Components() []component.Component
	RunPipeline(wg *sync.WaitGroup) error

	AddEngineRelay(engineChan chan core.InvariantInput) error
}

// pipeline ... Pipeline implementation
type pipeline struct {
	id  core.PUUID
	cfg *core.PipelineConfig

	aState ActivityState

	components []component.Component
}

// NewPipeline ... Initializer
func NewPipeline(cfg *core.PipelineConfig, pUUID core.PUUID, comps []component.Component) (Pipeline, error) {
	pl := &pipeline{
		cfg:        cfg,
		id:         pUUID,
		components: comps,
		aState:     Booting,
	}

	return pl, nil
}

// Config ... Returns pipeline config
func (pl *pipeline) Config() *core.PipelineConfig {
	return pl.cfg
}

// Components ... Returns slice of all constituent components
func (pl *pipeline) Components() []component.Component {
	return pl.components
}

// UUID ... Returns pipeline UUID
func (pl *pipeline) UUID() core.PUUID {
	return pl.id
}

// AddEngineRelay ... Adds a relay to the pipeline that forces it to send transformed invariant input
// to a risk engine
func (pl *pipeline) AddEngineRelay(engineChan chan core.InvariantInput) error {
	lastComponent := pl.components[0]
	eir := core.NewEngineRelay(pl.id, engineChan)

	logging.NoContext().Debug("Adding engine relay to pipeline",
		zap.String(core.CUUIDKey, lastComponent.UUID().String()),
		zap.String(core.PUUIDKey, pl.id.String()))

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
					zap.String(core.PUUIDKey, pl.id.String()))

			if err := c.EventLoop(); err != nil {
				logging.NoContext().Error("Obtained error from event loop", zap.Error(err),
					zap.String(core.CUUIDKey, c.UUID().String()),
					zap.String(core.PUUIDKey, pl.id.String()))
			}
		}(comp, wg)
	}

	return nil
}

// Close ... Closes all components in the pipeline
func (pl *pipeline) Close() error {
	for _, comp := range pl.components {
		if comp.ActivityState() != component.Terminated {
			logging.NoContext().
				Debug("Shutting down pipeline component",
					zap.String(core.CUUIDKey, comp.UUID().String()),
					zap.String(core.PUUIDKey, pl.id.String()))

			if err := comp.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
