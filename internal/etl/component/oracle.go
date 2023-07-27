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

// OracleDefinition ... Provides a generalized interface for developers to bind their own functionality to
type OracleDefinition interface {
	BackTestRoutine(ctx context.Context, componentChan chan core.TransitData,
		startHeight *big.Int, endHeight *big.Int) error
	ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error
}

// Oracle ... Component used to represent a data source reader; E.g, Eth block indexing, interval API polling
type Oracle struct {
	ctx context.Context

	definition    OracleDefinition
	oracleChannel chan core.TransitData

	wg *sync.WaitGroup

	*metaData
}

// NewOracle ... Initializer
func NewOracle(ctx context.Context, outType core.RegisterType,
	od OracleDefinition, opts ...Option) (Component, error) {
	o := &Oracle{
		ctx:           ctx,
		definition:    od,
		oracleChannel: core.NewTransitChannel(),
		wg:            &sync.WaitGroup{},

		metaData: newMetaData(core.Oracle, outType),
	}

	for _, opt := range opts {
		opt(o.metaData)
	}

	logging.WithContext(ctx).Info("Constructed component",
		zap.String(logging.CUUIDKey, o.metaData.id.String()))

	return o, nil
}

// Close ... This function is called at the end when processes related to oracle need to shut down
func (o *Oracle) Close() error {
	logging.WithContext(o.ctx).
		Info("Waiting for oracle definition go routines to finish",
			zap.String(logging.CUUIDKey, o.id.String()))
	o.closeChan <- killSig

	o.wg.Wait()
	logging.WithContext(o.ctx).Info("Oracle definition go routines have exited",
		zap.String(logging.CUUIDKey, o.id.String()))
	return nil
}

// EventLoop ... Component loop that actively waits and transits register data
// from a channel that the definition's read routine writes to
func (o *Oracle) EventLoop() error {
	// TODO(#24) - Add Internal Component Activity State Tracking

	logger := logging.WithContext(o.ctx)

	logger.Debug("Starting component event loop",
		zap.String(logging.CUUIDKey, o.id.String()))

	o.wg.Add(1)

	routineCtx, cancel := context.WithCancel(o.ctx)
	// o.emitStateChange(Live)

	// Spawn definition read routine
	go func() {
		defer o.wg.Done()
		if err := o.definition.ReadRoutine(routineCtx, o.oracleChannel); err != nil {
			logger.Error("Received error from read routine",
				zap.String(logging.CUUIDKey, o.id.String()),
				zap.Error(err))
		}
	}()

	for {
		select {
		case registerData := <-o.oracleChannel:
			logger.Debug("Sending data",
				zap.String(logging.CUUIDKey, o.id.String()))

			if err := o.egressHandler.Send(registerData); err != nil {
				logger.Error(transitErr, zap.String("ID", o.id.String()))
			}

			if o.egressHandler.PathEnd() {
				latency := float64(time.Since(registerData.OriginTS).Milliseconds())
				metrics.WithContext(o.ctx).
					RecordPipelineLatency(o.pUUID, latency)
			}

		case <-o.closeChan:
			logger.Debug("Received component shutdown signal",
				zap.String(logging.CUUIDKey, o.id.String()))

			// o.emitStateChange(Terminated)
			logger.Debug("Closing component channel and context",
				zap.String(logging.CUUIDKey, o.id.String()))
			close(o.oracleChannel)
			cancel() // End definition routine

			logger.Debug("Component shutdown success",
				zap.String(logging.CUUIDKey, o.id.String()))
			return nil
		}
	}
}
