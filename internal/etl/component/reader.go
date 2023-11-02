package component

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

// ReadRoutine ...
type ReadRoutine interface {
	Loop(ctx context.Context, componentChan chan core.TransitData) error
	Height() (*big.Int, error)
}

// ChainReader ...
type ChainReader struct {
	ctx context.Context

	routine       ReadRoutine
	oracleChannel chan core.TransitData

	wg *sync.WaitGroup

	*metaData
}

// NewReader ... Initializer
func NewReader(ctx context.Context, outType core.RegisterType,
	rr ReadRoutine, opts ...Option) (Component, error) {
	cr := &ChainReader{
		ctx:           ctx,
		routine:       rr,
		oracleChannel: core.NewTransitChannel(),
		wg:            &sync.WaitGroup{},

		metaData: newMetaData(core.Reader, outType),
	}

	for _, opt := range opts {
		opt(cr.metaData)
	}

	logging.WithContext(ctx).Info("Constructed component",
		zap.String(logging.CUUIDKey, cr.metaData.id.String()))

	return cr, nil
}

// Height ... Returns the current block height of the chain read routine
func (cr *ChainReader) Height() (*big.Int, error) {
	return cr.routine.Height()
}

// Close ... This function is called at the end when processes related to reader need to shut down
func (cr *ChainReader) Close() error {
	logging.WithContext(cr.ctx).
		Info("Waiting for reader definition go routines to finish",
			zap.String(logging.CUUIDKey, cr.id.String()))
	cr.closeChan <- killSig

	cr.wg.Wait()
	logging.WithContext(cr.ctx).Info("Reader definition go routines have exited",
		zap.String(logging.CUUIDKey, cr.id.String()))
	return nil
}

// EventLoop ...
func (cr *ChainReader) EventLoop() error {
	// TODO(#24) - Add Internal Component Activity State Tracking

	logger := logging.WithContext(cr.ctx)

	logger.Debug("Starting component event loop",
		zap.String(logging.CUUIDKey, cr.id.String()))

	cr.wg.Add(1)

	routineCtx, cancel := context.WithCancel(cr.ctx)
	// cr.emitStateChange(Live)

	// Spawn definition read routine
	go func() {
		defer cr.wg.Done()
		if err := cr.routine.Loop(routineCtx, cr.oracleChannel); err != nil {
			logger.Error("Received error from read routine",
				zap.String(logging.CUUIDKey, cr.id.String()),
				zap.Error(err))
		}
	}()

	for {
		select {
		case registerData := <-cr.oracleChannel:
			logger.Debug("Sending data",
				zap.String(logging.CUUIDKey, cr.id.String()))

			if err := cr.egressHandler.Send(registerData); err != nil {
				logger.Error(transitErr, zap.String("ID", cr.id.String()))
			}

			if cr.egressHandler.PathEnd() {
				latency := float64(time.Since(registerData.OriginTS).Milliseconds())
				metrics.WithContext(cr.ctx).
					RecordPipelineLatency(cr.pUUID, latency)
			}

		case <-cr.closeChan:
			logger.Debug("Received component shutdown signal",
				zap.String(logging.CUUIDKey, cr.id.String()))

			// cr.emitStateChange(Terminated)
			logger.Debug("Closing component channel and context",
				zap.String(logging.CUUIDKey, cr.id.String()))
			close(cr.oracleChannel)
			cancel() // End definition routine

			logger.Debug("Component shutdown success",
				zap.String(logging.CUUIDKey, cr.id.String()))
			return nil
		}
	}
}
