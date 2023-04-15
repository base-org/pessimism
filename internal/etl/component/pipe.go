package component

import (
	"context"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// TransformFunc ... Generic transformation function
type TransformFunc func(data core.TransitData) ([]core.TransitData, error)

// Pipe ... Component used to represent any arbitrary computation; pipes must can read from all component types
// E.G. (ORACLE || CONVEYOR || PIPE) -> PIPE

type Pipe struct {
	ctx    context.Context
	inType core.RegisterType

	tform TransformFunc

	*metaData
}

// NewPipe ... Initializer
func NewPipe(ctx context.Context, tform TransformFunc, inType core.RegisterType,
	outType core.RegisterType, opts ...Option) (Component, error) {
	// TODO - Validate inTypes size

	pipe := &Pipe{
		ctx:    ctx,
		tform:  tform,
		inType: inType,

		metaData: &metaData{
			id:             core.NilCompID(),
			cType:          core.Pipe,
			egressHandler:  newEgressHandler(),
			ingressHandler: newIngressHandler(),
			state:          Inactive,
			output:         outType,
			RWMutex:        &sync.RWMutex{},
		},
	}

	if err := pipe.createIngress(inType); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(pipe.metaData)
	}

	log.Printf("[%s] Constructed component", pipe.metaData.id.String())
	return pipe, nil
}

func (p *Pipe) Close() {
}

// EventLoop ... Driver loop for component that actively subscribes
// to an input channel where transit data is read, transformed, and transitte
// to downstream components
func (p *Pipe) EventLoop() error {
	logging.WithContext(p.ctx).Info("Starting event loop",
		zap.String("ID", p.id.String()),
	)

	p.metaData.state = Live
	inChan, err := p.GetIngress(p.inType)
	if err != nil {
		return err
	}

	for {
		select {
		case inputData := <-inChan:
			outputData, err := p.tform(inputData)
			if err != nil {
				// TODO - Introduce prometheus call here
				// TODO - Introduce go standard logger (I,E. zap) debug call
				log.Printf("%e", err)
				continue
			}

			if len(outputData) > 0 {
				log.Printf("[%s][%s] Received output data: %s", p.id, p.cType, outputData[0].Type)
			}

			logging.WithContext(p.ctx).Debug("Sending data batch",
				zap.String("From", p.id.String()))

			if err := p.egressHandler.SendBatch(outputData); err != nil {
				log.Printf(transitErr, p.id, p.cType, err.Error())
			}

		// Manager is telling us to shutdown
		case <-p.ctx.Done():
			return nil
		}
	}
}
