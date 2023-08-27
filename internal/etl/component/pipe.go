package component

import (
	"context"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

type PipeDefinition interface {
	Transform(ctx context.Context, data core.TransitData) ([]core.TransitData, error)
}

// Pipe ... Component used to represent any arbitrary computation; pipes must can read from all component types
// E.G. (ORACLE || CONVEYOR || PIPE) -> PIPE

type Pipe struct {
	ctx    context.Context
	inType core.RegisterType

	def PipeDefinition

	*metaData
}

// NewPipe ... Initializer
func NewPipe(ctx context.Context, pd PipeDefinition, inType core.RegisterType,
	outType core.RegisterType, opts ...Option) (Component, error) {
	// TODO - Validate inTypes size

	pipe := &Pipe{
		ctx:    ctx,
		def:    pd,
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

// Close ... Shuts down component by emitting a kill signal to a close channel
func (p *Pipe) Close() error {
	p.closeChan <- killSig

	return nil
}

// EventLoop ... Driver loop for component that actively subscribes
// to an input channel where transit data is read, transformed, and transited
// to downstream components
func (p *Pipe) EventLoop() error {
	logger := logging.WithContext(p.ctx)

	logger.Info("Starting event loop",
		zap.String("ID", p.id.String()),
	)

	inChan, err := p.GetIngress(p.inType)
	if err != nil {
		return err
	}

	for {
		select {
		case inputData := <-inChan:
			// 1. Transform the input data to some output data
			// NOTE ... Continuing to process data even if there is an error
			// is done in the event of partial processing where some data
			// may be successfully processed and some may not
			outputData, err := p.def.Transform(p.ctx, inputData)
			if err != nil {
				// TODO - Introduce metrics service (`prometheus`) call
				logger.Error(err.Error(), zap.String("ID", p.id.String()))
			}

			// 2. Determine if component is at the end of a pipeline, emit metrics if so
			if p.egressHandler.PathEnd() {
				latency := float64(time.Since(inputData.OriginTS).Milliseconds())

				metrics.WithContext(p.ctx).
					RecordPipelineLatency(p.pUUID,
						latency)
			}

			// 3. Verify that some output data was produced and continue if not
			length := len(outputData)
			logger.Debug("Received transformation output data",
				zap.String("ID", p.id.String()),
				zap.Int("Length", length))

			if length == 0 {
				continue
			}

			// 4. Batch send output data to subscribed downstream components
			logger.Debug("Sending data batch",
				zap.String("ID", p.id.String()),
				zap.String("Type", p.OutputType().String()))

			if err := p.egressHandler.SendBatch(outputData); err != nil {
				logger.Error(transitErr, zap.String("ID", p.id.String()))
			}

		// Manager is telling us to shutdown
		case <-p.closeChan:
			logger.Debug("Received component shutdown signal",
				zap.String("ID", p.id.String()))

			// p.emitStateChange(Terminated)

			return nil
		}
	}
}
