package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/config"
)

// OracleType ...
type OracleType = string

const (
	// BackTestOracle ... Represents an oracle used for backtesting some invariant
	BacktestOracle OracleType = "backtest"
	// LiveOracle ... Represents an oracle used for powering some live invariant
	LiveOracle OracleType = "live"
)

// OracleDefinition ... Provides a generalized interface for developers to bind their own functionality to
type OracleDefinition interface {
	ConfigureRoutine() error
	BackTestRoutine(ctx context.Context, componentChan chan models.TransitData) error
	ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error
}

// OracleConstructor ... Type declaration for what a registry oracle constructor function must adhere to
type OracleConstructor = func(ctx context.Context, ot OracleType, cfg *config.OracleConfig) PipelineComponent

// OracleOption ...
type OracleOption = func(*Oracle)

// Oracle ... Component used to represent a data source reader; E.g, Eth block indexing, interval API polling
type Oracle struct {
	ctx context.Context

	od OracleDefinition
	ot OracleType

	*OutputRouter
}

// Type ... Returns the pipeline component type
func (o *Oracle) Type() models.ComponentType {
	return models.Oracle
}

// NewOracle ... Initializer
func NewOracle(ctx context.Context, ot OracleType, od OracleDefinition, opts ...OracleOption) PipelineComponent {
	o := &Oracle{
		ctx:          ctx,
		od:           od,
		ot:           ot,
		OutputRouter: NewOutputRouter(),
	}

	for _, opt := range opts {
		opt(o)
	}

	od.ConfigureRoutine()
	return o
}

// EventLoop ... Component loop that actively waits and transits register data from a channel that the definition's read routine writes to
func (o *Oracle) EventLoop() error {

	oracleChannel := make(chan models.TransitData)

	// Spawn read routine process
	// TODO - Consider higher order concurrency injection; ie waitgroup, routine management
	go o.od.ReadRoutine(o.ctx, oracleChannel)

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
