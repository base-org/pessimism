package pipeline

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Pipeline ... Pipeline interface
type Pipeline interface {
	BlockHeight() (*big.Int, error)
	Config() *core.PipelineConfig
	Components() []component.Component
	UUID() core.PUUID
	State() ActivityState

	Close() error
	Run(wg *sync.WaitGroup)
	AddEngineRelay(engineChan chan core.HeuristicInput) error
}

// pipeline ... Pipeline implementation
type pipeline struct {
	id  core.PUUID
	cfg *core.PipelineConfig

	state ActivityState

	components []component.Component
}

// NewPipeline ... Initializer
func NewPipeline(cfg *core.PipelineConfig, pUUID core.PUUID, comps []component.Component) (Pipeline, error) {
	if len(comps) == 0 {
		return nil, fmt.Errorf(emptyPipelineError)
	}

	pl := &pipeline{
		cfg:        cfg,
		id:         pUUID,
		components: comps,
		state:      INACTIVE,
	}

	return pl, nil
}

// State ... Returns pipeline state
func (pl *pipeline) State() ActivityState {
	return pl.state
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

func (pl *pipeline) BlockHeight() (*big.Int, error) {
	comp := pl.components[len(pl.components)-1]
	cr, ok := comp.(*component.ChainReader)
	if !ok {
		return nil, fmt.Errorf("could not cast component to chain reader")
	}

	return cr.Height()
}

// AddEngineRelay ... Adds a relay to the pipeline that forces it to send transformed heuristic input
// to a risk engine
func (pl *pipeline) AddEngineRelay(engineChan chan core.HeuristicInput) error {
	lastComponent := pl.components[0]
	eir := core.NewEngineRelay(pl.id, engineChan)

	logging.NoContext().Debug("Adding engine relay to pipeline",
		zap.String(logging.CUUIDKey, lastComponent.UUID().String()),
		zap.String(logging.PUUIDKey, pl.id.String()))

	return lastComponent.AddRelay(eir)
}

// Run  ... Spawns and manages component event loops
// for some pipeline
func (pl *pipeline) Run(wg *sync.WaitGroup) {
	for _, comp := range pl.components {
		wg.Add(1)

		go func(c component.Component, wg *sync.WaitGroup) {
			defer wg.Done()

			logging.NoContext().
				Debug("Attempting to start component event loop",
					zap.String(logging.CUUIDKey, c.UUID().String()),
					zap.String(logging.PUUIDKey, pl.id.String()))

			if err := c.EventLoop(); err != nil {
				// NOTE - Consider killing the entire pipeline if one component fails
				// Otherwise dangling components will be left in a running state
				logging.NoContext().Error("Obtained error from event loop", zap.Error(err),
					zap.String(logging.CUUIDKey, c.UUID().String()),
					zap.String(logging.PUUIDKey, pl.id.String()))
				pl.state = CRASHED
			}
		}(comp, wg)
	}

	pl.state = ACTIVE
}

// Close ... Closes all components in the pipeline
func (pl *pipeline) Close() error {
	for _, comp := range pl.components {
		if comp.ActivityState() != component.Terminated {
			logging.NoContext().
				Debug("Shutting down pipeline component",
					zap.String(logging.CUUIDKey, comp.UUID().String()),
					zap.String(logging.PUUIDKey, pl.id.String()))

			if err := comp.Close(); err != nil {
				return err
			}
		}
	}
	pl.state = TERMINATED
	return nil
}
