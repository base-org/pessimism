package etl

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Process path
type Path interface {
	BlockHeight() (*big.Int, error)
	Config() *core.PathConfig
	Processes() []process.Process
	UUID() core.PathID
	State() ActivityState

	Close() error
	Run(wg *sync.WaitGroup)
	AddEngineRelay(engineChan chan core.HeuristicInput) error
}

type path struct {
	id  core.PathID
	cfg *core.PathConfig

	state ActivityState

	processes []process.Process
}

// NewPath ... Initializer
func NewPath(cfg *core.PathConfig, PathID core.PathID, procs []process.Process) (Path, error) {
	if len(procs) == 0 {
		return nil, fmt.Errorf(emptyPathError)
	}

	p := &path{
		cfg:       cfg,
		id:        PathID,
		processes: procs,
		state:     INACTIVE,
	}

	return p, nil
}

func (p *path) State() ActivityState {
	return p.state
}

func (p *path) Config() *core.PathConfig {
	return p.cfg
}

func (p *path) Processes() []process.Process {
	return p.processes
}

func (p *path) UUID() core.PathID {
	return p.id
}

func (p *path) BlockHeight() (*big.Int, error) {
	// We assume that all paths have an oracle as their last process
	comp := p.processes[len(p.processes)-1]
	cr, ok := comp.(*process.ChainReader)
	if !ok {
		return nil, fmt.Errorf("could not cast process to chain reader")
	}

	return cr.Height()
}

// AddEngineRelay ... Adds a relay to the path that forces it to send transformed heuristic input
// to a risk engine
func (p *path) AddEngineRelay(engineChan chan core.HeuristicInput) error {
	lastProcess := p.processes[0]
	eir := core.NewEngineRelay(p.id, engineChan)

	logging.NoContext().Debug("Adding engine relay to path",
		zap.String(logging.Process, lastProcess.ID().String()),
		zap.String(logging.Path, p.id.String()))

	return lastProcess.AddEngineRelay(eir)
}

// Run  ... Spawns process event loops
func (path *path) Run(wg *sync.WaitGroup) {
	for _, p := range path.processes {
		wg.Add(1)

		go func(p process.Process, wg *sync.WaitGroup) {
			defer wg.Done()

			logging.NoContext().
				Debug("Starting process",
					zap.String(logging.Process, p.ID().String()),
					zap.String(logging.Path, p.ID().String()))

			if err := p.EventLoop(); err != nil {
				// NOTE - Consider killing the entire path if one process fails
				// Otherwise dangling processes will be left in a running state
				logging.NoContext().Error("Obtained error from event loop", zap.Error(err),
					zap.String(logging.Process, p.ID().String()),
					zap.String(logging.Path, p.ID().String()))
				path.state = CRASHED
			}
		}(p, wg)
	}

	path.state = ACTIVE
}

// Close ... Closes all processes in the path
func (p *path) Close() error {
	for _, p := range p.processes {
		if p.ActivityState() != process.Terminated {
			logging.NoContext().
				Debug("Shutting down path process",
					zap.String(logging.Process, p.ID().String()),
					zap.String(logging.Path, p.ID().String()))

			if err := p.Close(); err != nil {
				return err
			}
		}
	}
	p.state = TERMINATED
	return nil
}
