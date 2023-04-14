package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type PipeOption func(*Pipe)

func WithRouter(router *OutputRouter) PipeOption {
	return func(p *Pipe) {
		p.OutputRouter = router
	}
}

// TransformFunc ... Generic transformation function
type TranformFunc func(data models.TransitData) ([]models.TransitData, error)

// Pipe ... Component used to represent any arbitrary computation; pipes must always read from an existing component
// E.G, (ORACLE || CONVEYOR || PIPE) -> PIPE

type Pipe struct {
	ctx   context.Context
	tform TranformFunc

	// Channel that a pipe is subscribed to for new data events
	inputChan chan models.TransitData

	*OutputRouter
}

// NewPipe ... Initializer
func NewPipe(ctx context.Context, tform TranformFunc,
	inputChan chan models.TransitData, opts ...PipeOption) (Component, error) {
	log := ctxzap.Extract(ctx)
	log.Info("Constructing new component pipe")

	router, err := NewOutputRouter()
	if err != nil {
		return nil, err
	}

	pipe := &Pipe{
		ctx:          ctx,
		tform:        tform,
		inputChan:    inputChan,
		OutputRouter: router,
	}

	for _, opt := range opts {
		opt(pipe)
	}

	return pipe, nil
}

// Type ... Returns component type
func (p *Pipe) Type() models.ComponentType {
	return models.Pipe
}

func (p *Pipe) Close() {
}

// EventLoop ... Driver loop for component that actively subscribes
// to an input channel where transit data is read, transformed, and transitte
// to downstream components
func (p *Pipe) EventLoop() error {
	log := ctxzap.Extract(p.ctx)
	for {
		select {
		// Input has been fed to the component
		case inputData := <-p.inputChan:
			log.Info("Got input data")
			outputData, err := p.tform(inputData)
			if err != nil {
				// TODO - Introduce prometheus call here
				// TODO - Introduce go standard logging (I,E. zap) debug call
				log.Error("error transforming", zap.Error(err))
				continue
			}

			log.Info("Transiting output")
			p.OutputRouter.TransitOutputs(outputData)

		// Manager is telling us to shutdown
		case <-p.ctx.Done():
			return nil
		}
	}
}
