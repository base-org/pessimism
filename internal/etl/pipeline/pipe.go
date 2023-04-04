package pipeline

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
	id  models.ComponentID

	tform TranformFunc

	// Channel that a pipe is subscribed to for new data events
	inputChan chan models.TransitData

	*metaData
}

// NewPipe ... Initializer
func NewPipe(ctx context.Context, tform TranformFunc, opts ...ComponentOption) (Component, error) {
	log.Print("Constructing new component pipe ")

	router, err := newRouter()
	if err != nil {
		return nil, err
	}

	pipe := &Pipe{
		ctx:       ctx,
		tform:     tform,
		inputChan: models.NewTransitChannel(),

		metaData: &metaData{
			id:     uuid.New(),
			cType:  models.Pipe,
			router: router,
		},
	}

	for _, opt := range opts {
		opt(pipe.metaData)
	}

	return pipe, nil
}

// EventLoop ... Driver loop for component that actively subscribes
// to an input channel where transit data is read, transformed, and transitte
// to downstream components
func (p *Pipe) EventLoop() error {
	for {
		select {
		// Input has been fed to the component
		case inputData := <-p.inputChan:
			log.Printf("Got input data")
			outputData, err := p.tform(inputData)
			if err != nil {
				// TODO - Introduce prometheus call here
				// TODO - Introduce go standard logger (I,E. zap) debug call
				log.Printf("%e", err)
				continue
			}

			log.Printf("Transiting output")
			p.router.TransitOutputs(outputData)

		// Manager is telling us to shutdown
		case <-p.ctx.Done():
			return nil
		}
	}
}

func (p *Pipe) EntryPoints() []chan models.TransitData {
	return []chan models.TransitData{p.inputChan}
}
