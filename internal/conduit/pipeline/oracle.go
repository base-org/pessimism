package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/config"
)

type OracleType = string

const (
	BacktestOracle OracleType = "backtest"
	LiveOracle     OracleType = "live"
)

// OracleDefinition provides a plug & play
type OracleDefinition interface {
	ConfigureRoutine() error
	ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error
}

type OracleConstructor = func(ctx context.Context, ot OracleType, cfg *config.OracleConfig) PipelineComponent

type OracleOption = func(*Oracle)

type Oracle struct {
	ctx context.Context

	od OracleDefinition

	*OutputRouter
}

func (o *Oracle) Type() models.ComponentType {
	return models.Oracle
}

func NewOracle(ctx context.Context, ot OracleType, od OracleDefinition, opts ...OracleOption) PipelineComponent {
	if ot == LiveOracle {
		o := &Oracle{
			ctx:          ctx,
			od:           od,
			OutputRouter: NewOutputRouter(),
		}

		for _, opt := range opts {
			opt(o)
		}

		od.ConfigureRoutine()
		return o

	} else if ot == BacktestOracle {

	}

	return nil

}

func (o *Oracle) EventLoop() error {

	oracleChannel := make(chan models.TransitData)

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
