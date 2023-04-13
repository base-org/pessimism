package pipeline

import (
	"context"
	"log"
	"math/big"
	"sync"

	"github.com/base-org/pessimism/internal/conduit/models"
)

// OracleDefinition ... Provides a generalized interface for developers to bind their own functionality to
type OracleDefinition interface {
	ConfigureRoutine() error
	BackTestRoutine(ctx context.Context, componentChan chan models.TransitData, startHeight *big.Int,
		endHeight *big.Int) error
	ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error
}

// OracleOption ...
type OracleOption = func(*Oracle)

// Oracle ... Component used to represent a data source reader; E.g, Eth block indexing, interval API polling
type Oracle struct {
	ctx context.Context

	od        OracleDefinition
	ot        OracleType
	waitGroup *sync.WaitGroup

	*OutputRouter
}

// Type ... Returns the pipeline component type
func (o *Oracle) Type() models.ComponentType {
	return models.Oracle
}

// NewOracle ... Initializer
func NewOracle(ctx context.Context, ot OracleType,
	od OracleDefinition, opts ...OracleOption) (Component, error) {
	router, err := NewOutputRouter()
	if err != nil {
		return nil, err
	}

	o := &Oracle{
		ctx:          ctx,
		od:           od,
		ot:           ot,
		waitGroup:    &sync.WaitGroup{},
		OutputRouter: router,
	}

	for _, opt := range opts {
		opt(o)
	}

	if cfgErr := od.ConfigureRoutine(); cfgErr != nil {
		return nil, cfgErr
	}

	return o, nil
}

func (o *Oracle) Close() {
	log.Printf("Waiting for oracle goroutines to be done.")
	o.waitGroup.Wait()
	log.Printf("Oracle goroutines have exited.")
}

// EventLoop ... Component loop that actively waits and transits register data
// from a channel that the definition's read routine writes to
func (o *Oracle) EventLoop() error {
	oracleChannel := make(chan models.TransitData)

	// Spawn read routine process
	o.waitGroup.Add(1)
	go func() {
		defer o.waitGroup.Done()
		if err := o.od.ReadRoutine(o.ctx, oracleChannel); err != nil {
			log.Printf("Received error from read routine %s", err.Error())
		}
	}()

	for {
		select {
		case registerData := <-oracleChannel:
			o.OutputRouter.TransitOutput(registerData)

		case <-o.ctx.Done():
			close(oracleChannel)
			return nil
		}
	}
}
