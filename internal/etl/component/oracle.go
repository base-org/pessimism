package component

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// OracleDefinition ... Provides a generalized interface for developers to bind their own functionality to
type OracleDefinition interface {
	ConfigureRoutine() error
	BackTestRoutine(ctx context.Context, componentChan chan core.TransitData, startHeight *big.Int, endHeight *big.Int) error
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

		metaData: &metaData{
			id:             core.NilCompID(),
			cType:          core.Oracle,
			egressHandler:  newEgressHandler(),
			ingressHandler: newIngressHandler(),
			state:          Inactive,
			output:         outType,
			RWMutex:        &sync.RWMutex{},
		},
	}

	if od == nil {
		return nil, fmt.Errorf("Received nil value for OracleDefinition")
	}

	for _, opt := range opts {
		opt(o.metaData)
	}

	if cfgErr := od.ConfigureRoutine(); cfgErr != nil {
		return nil, cfgErr
	}

	logging.WithContext(ctx).Info("Constructed component",
		zap.String("ID", o.metaData.id.String()))

	return o, nil
}

// TODO (#22) : Add closure logic to all component types

// Close ... This function is called at the end when processes related to oracle need to shut down
func (o *Oracle) Close() {
	logging.WithContext(o.ctx).Info("Waiting for oracle goroutines to be done.")
	o.wg.Wait()
	logging.WithContext(o.ctx).Info("Oracle goroutines have exited.")
}

// EventLoop ... Component loop that actively waits and transits register data
// from a channel that the definition's read routine writes to
func (o *Oracle) EventLoop() error {
	// TODO - Introduce backfill check so that state can be "syncing"
	o.state = Live
	log.Printf("[%s][%s] Starting event loop", o.id, o.cType)

	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		if err := o.definition.ReadRoutine(o.ctx, o.oracleChannel); err != nil {
			logging.WithContext(o.ctx).Error("Received error from read routine", zap.Error(err))
		}
	}()

	for {
		select {
		case registerData := <-o.oracleChannel:
			logging.WithContext(o.ctx).Debug("Sending data",
				zap.String("From", o.id.String()))

			if err := o.egressHandler.Send(registerData); err != nil {
				logging.WithContext(o.ctx).Error(
					fmt.Sprintf(transitErr, o.id, o.cType, err.Error()))
			}

		case <-o.ctx.Done():
			close(o.oracleChannel)

			return nil
		}
	}
}
