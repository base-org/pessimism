package pipeline

import (
	"context"
	"log"

	"github.com/base-org/pessimism/internal/conduit/models"
)

type PipeOption func(*Pipe)

func WithRouter(router *OutputRouter) PipeOption {
	return func(p *Pipe) {
		p.router = router
	}
}

// TransformFunc ...
type TranformFunc func(data models.TransitData) (*models.TransitData, error)

type Pipe struct {
	ctx   context.Context
	tform TranformFunc

	inputChan chan models.TransitData
	router    *OutputRouter
}

func NewPipe(ctx context.Context, tform TranformFunc, inputChan chan models.TransitData, opts ...PipeOption) PipelineComponent {
	log.Print("Constructing new component pipe ")

	pipe := &Pipe{
		ctx:       ctx,
		tform:     tform,
		inputChan: inputChan,
		router:    NewOutputRouter(),
	}

	for _, opt := range opts {
		opt(pipe)
	}

	return pipe
}

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
			p.router.TransitOutput(*outputData)

		// Manager is telling us to shutdown
		case <-p.ctx.Done():
			return nil
		}

	}

}
