package component

import (
	"context"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

// OracleDefinition ... Provides a generalized interface for developers to bind their own functionality to
type OracleDefinition interface {
	ConfigureRoutine() error
	BackTestRoutine(ctx context.Context, componentChan chan core.TransitData) error
	ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error
}

// Oracle ... Component used to represent a data source reader; E.g, Eth block indexing, interval API polling
type Oracle struct {
	ctx context.Context

	definition    OracleDefinition
	oracleType    core.PipelineType
	oracleChannel chan core.TransitData

	*metaData
}

// NewOracle ... Initializer
func NewOracle(ctx context.Context, pt core.PipelineType, outType core.RegisterType,
	od OracleDefinition, opts ...Option) (Component, error) {
	router, err := newRouter()
	if err != nil {
		return nil, err
	}

	o := &Oracle{
		ctx:           ctx,
		definition:    od,
		oracleType:    pt,
		oracleChannel: core.NewTransitChannel(),

		metaData: &metaData{
			id:      core.NilCompID(),
			cType:   core.Oracle,
			router:  router,
			ingress: newIngress(),
			state:   Inactive,
			output:  outType,
			RWMutex: &sync.RWMutex{},
		},
	}

	for _, opt := range opts {
		opt(o.metaData)
	}

	if cfgErr := od.ConfigureRoutine(); cfgErr != nil {
		return nil, cfgErr
	}

	log.Printf("[%s] Constructed component", o.metaData.id.String())
	return o, nil
}

// EventLoop ... Component loop that actively waits and transits register data
// from a channel that the definition's read routine writes to
func (o *Oracle) EventLoop() error {
	o.RWMutex.Lock()
	defer o.RWMutex.Unlock()
	// TODO - Introduce backfill check so that state can be "syncing"
	o.state = Live
	log.Printf("[%s][%s] Starting event loop", o.id, o.cType)
	// Spawn read routine process
	// TODO - Consider higher order concurrency injection; ie waitgroup, routine management
	go func(ch chan core.TransitData) {
		if err := o.definition.ReadRoutine(o.ctx, ch); err != nil {
			log.Printf("[%s][%s] Received error from read routine %s", o.id, o.cType, err.Error())
		}
	}(o.oracleChannel)

	for {
		select {
		case registerData := <-o.oracleChannel:
			log.Printf("")
			if err := o.router.TransitOutput(registerData); err != nil {
				log.Printf(transitErr, o.id, o.cType, err.Error())
			}

		case <-o.ctx.Done():
			close(o.oracleChannel)

			return nil
		}
	}
}
