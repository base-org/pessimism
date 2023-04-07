package component

import (
	"context"
	"log"

	"github.com/base-org/pessimism/internal/models"
	"github.com/google/uuid"
)

// TransformFunc ... Generic transformation function
type TranformFunc func(data models.TransitData) ([]models.TransitData, error)

// Pipe ... Component used to represent any arbitrary computation; pipes must always read from an existing component
// E.G, (ORACLE || CONVEYOR || PIPE) -> PIPE

type Pipe struct {
	ctx context.Context
	id  models.ID

	inType models.RegisterType
	tform  TranformFunc

	*metaData
}

// NewPipe ... Initializer
func NewPipe(ctx context.Context, tform TranformFunc, inType models.RegisterType, opts ...Option) (Component, error) {
	log.Print("Constructing new component pipe ")

	// TODO - Validate inTypes size

	router, err := newRouter()
	if err != nil {
		return nil, err
	}

	pipe := &Pipe{
		ctx:    ctx,
		tform:  tform,
		inType: inType,

		metaData: &metaData{
			id:      uuid.New(),
			cType:   models.Pipe,
			router:  router,
			ingress: newIngress(),
			state:   Inactive,
		},
	}

	log.Printf("Creating entry point for %s", inType)
	pipe.CreateEntryPoint(inType)

	for _, opt := range opts {
		opt(pipe.metaData)
	}

	return pipe, nil
}

// EventLoop ... Driver loop for component that actively subscribes
// to an input channel where transit data is read, transformed, and transitte
// to downstream components
func (p *Pipe) EventLoop() error {
	inChan, err := p.GetEntryPoint(p.inType)
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

			if err := p.router.TransitOutputs(outputData); err != nil {
				log.Printf(transitErr, p.id, p.cType, err.Error())
			}

		// Manager is telling us to shutdown
		case <-p.ctx.Done():
			return nil
		}
	}
}
