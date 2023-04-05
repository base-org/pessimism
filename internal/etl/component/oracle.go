package component

import (
	"context"
	"log"

	"github.com/base-org/pessimism/internal/models"
	"github.com/google/uuid"
)

// OracleDefinition ... Provides a generalized interface for developers to bind their own functionality to
type OracleDefinition interface {
	ConfigureRoutine() error
	BackTestRoutine(ctx context.Context, componentChan chan models.TransitData) error
	ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error
}

// Oracle ... Component used to represent a data source reader; E.g, Eth block indexing, interval API polling
type Oracle struct {
	ctx context.Context
	id  models.ComponentID

	definition    OracleDefinition
	oracleType    models.PipelineType
	oracleChannel chan models.TransitData

	*metaData
}

// NewOracle ... Initializer
func NewOracle(ctx context.Context, pt models.PipelineType,
	od OracleDefinition, opts ...Option) (Component, error) {
	router, err := newRouter()
	if err != nil {
		return nil, err
	}

	o := &Oracle{
		ctx:           ctx,
		definition:    od,
		oracleType:    pt,
		oracleChannel: models.NewTransitChannel(),

		metaData: &metaData{
			id:     uuid.New(),
			cType:  models.Oracle,
			router: router,
		},
	}

	for _, opt := range opts {
		opt(o.metaData)
	}

	if cfgErr := od.ConfigureRoutine(); cfgErr != nil {
		return nil, cfgErr
	}

	return o, nil
}

// EventLoop ... Component loop that actively waits and transits register data
// from a channel that the definition's read routine writes to
func (o *Oracle) EventLoop() error {
	log.Printf("[%s][%s] Starting event loop", o.id, o.cType)
	// Spawn read routine process
	// TODO - Consider higher order concurrency injection; ie waitgroup, routine management
	go func(ch chan models.TransitData) {
		if err := o.definition.ReadRoutine(o.ctx, ch); err != nil {
			log.Printf("[%s][%s] Received error from read routine %s", o.id, o.cType, err.Error())
		}
	}(o.oracleChannel)

	for {
		select {
		case registerData := <-o.oracleChannel:
			if err := o.router.TransitOutput(registerData); err != nil {
				log.Printf(transitErr, o.id, o.cType, err.Error())
			}

		case <-o.ctx.Done():
			close(o.oracleChannel)

			return nil
		}
	}
}

func (o *Oracle) EntryPoints() []chan models.TransitData {
	return []chan models.TransitData{o.oracleChannel}
}
