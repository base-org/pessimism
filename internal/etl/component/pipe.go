package component

import (
	"context"

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

		metaData: newMetaData(core.Pipe, outType),
	}

	if err := pipe.createIngress(inType); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(pipe.metaData)
	}

	return pipe, nil
}

// TODO(#22): Add closure logic to all component types
func (p *Pipe) Close() {
}

// EventLoop ... Driver loop for component that actively subscribes
// to an input channel where transit data is read, transformed, and transitte
// to downstream components
func (p *Pipe) EventLoop() error {
	logger := logging.WithContext(p.ctx)

	logger.Info("Starting event loop",
		zap.String("ID", p.id.String()),
	)

	p.emitStateChange(Live)

	inChan, err := p.GetIngress(p.inType)
	if err != nil {
		return err
	}

	for {
		select {
		case inputData := <-inChan:
			outputData, err := p.tform(inputData)
			if err != nil {
				// TODO - Introduce metrics service (`prometheus`) call
				logger.Error(err.Error(), zap.String("ID", p.id.String()))
				continue
			}

			if length := len(outputData); length > 0 {
				logger.Debug("Received tranformation output data",
					zap.String("ID", p.id.String()),
					zap.Int("Length", length))
			} else {
				logger.Debug("Received output data of length 0",
					zap.String("ID", p.id.String()))
				continue
			}

			logger.Debug("Sending data batch",
				zap.String("ID", p.id.String()),
				zap.String("Type", p.OutputType().String()))

			if err := p.egressHandler.SendBatch(outputData); err != nil {
				logger.Error(transitErr, zap.String("ID", p.id.String()))
			}

		// Manager is telling us to shutdown
		case <-p.ctx.Done():
			p.emitStateChange(Terminated)

			return nil
		}
	}
}
