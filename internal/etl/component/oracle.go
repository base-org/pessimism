package component

import (
	"context"
	"math/big"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// OracleDefinition ... Provides a generalized interface for developers to bind their own functionality to
type OracleDefinition interface {
	ConfigureRoutine() error
	BackTestRoutine(ctx context.Context, componentChan chan core.TransitData,
		startHeight *big.Int, endHeight *big.Int) error
	ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error
}

// Oracle ... Component used to represent a data source reader; E.g, Eth block indexing, interval API polling
type Oracle struct {
	ctx context.Context

	definition    OracleDefinition
	oracleType    core.PipelineType
	oracleChannel chan core.TransitData

	wg *sync.WaitGroup

	*metaData
}

// NewOracle ... Initializer
func NewOracle(ctx context.Context, pt core.PipelineType, outType core.RegisterType,
	od OracleDefinition, opts ...Option) (Component, error) {
	o := &Oracle{
		ctx:           ctx,
		definition:    od,
		oracleType:    pt,
		oracleChannel: core.NewTransitChannel(),
		wg:            &sync.WaitGroup{},

		metaData: newMetaData(core.Oracle, outType),
	}

	for _, opt := range opts {
		opt(o.metaData)
	}

	if cfgErr := od.ConfigureRoutine(); cfgErr != nil {
		return nil, cfgErr
	}

	logging.WithContext(ctx).Info("Constructed component",
		zap.String("cuuid", o.metaData.id.String()))

	return o, nil
}

// Close ... This function is called at the end when processes related to oracle need to shut down
func (o *Oracle) Close() error {
	logging.WithContext(o.ctx).
		Info("Waiting for oracle definition go routines to finish",
			zap.String("cuuid", o.id.String()))
	o.closeChan <- KillSignal

	o.wg.Wait()
	logging.WithContext(o.ctx).Info("Oracle definition go routines have exited",
		zap.String("cuuid", o.id.String()))
	return nil
}

// EventLoop ... Component loop that actively waits and transits register data
// from a channel that the definition's read routine writes to
func (o *Oracle) EventLoop() error {
	// TODO(#24) - Add Internal Component Activity State Tracking

	logger := logging.WithContext(o.ctx)

	logger.Debug("Starting component event loop",
		zap.String("cuuid", o.id.String()))

	o.wg.Add(1)

	routineCtx, cancel := context.WithCancel(o.ctx)
	// o.emitStateChange(Live)

	// Spawn definition read routine
	go func() {
		defer o.wg.Done()
		if err := o.definition.ReadRoutine(routineCtx, o.oracleChannel); err != nil {
			logger.Error("Received error from read routine",
				zap.String("cuuid", o.id.String()),
				zap.Error(err))
		}
	}()

	for {
		select {
		case registerData := <-o.oracleChannel:
			logger.Debug("Sending data",
				zap.String("cuuid", o.id.String()))

			if err := o.egressHandler.Send(registerData); err != nil {
				logger.Error(transitErr, zap.String("ID", o.id.String()))
			}

		case <-o.closeChan:
			logger.Debug("Received component shutdown signal",
				zap.String("cuuid", o.id.String()))

			// o.emitStateChange(Terminated)
			logger.Debug("Closing component channel",
				zap.String("cuuid", o.id.String()))
			close(o.oracleChannel)
			logger.Debug("Closed component channel",
				zap.String("cuuid", o.id.String()))
			cancel() // End definition routine

			logger.Debug("Component shutdown success",
				zap.String("cuuid", o.id.String()))
			return nil
		}
	}
}
