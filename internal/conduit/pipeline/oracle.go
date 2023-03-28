package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/conduit/models"
)

type OracleType = string

const (
	Backtest OracleType = "backtest"
	Live     OracleType = "live"
)

type ConnectionType = int

type OracleDefinition interface {
	ConfigureRoutine() error
	ReadRoutine(context.Context, chan models.TransitData) error
}

type LiveOracleOption = func(*LiveOracle)

type LiveOracle struct {
	ct  ConnectionType
	ctx context.Context

	od OracleDefinition

	Routing
}

func (lo *LiveOracle) Type() models.ComponentType {
	return models.Oracle
}

func NewOracle(ctx context.Context, ot OracleType, od OracleDefinition, opts ...LiveOracleOption) PipelineComponent {
	if ot == Live {
		lo := &LiveOracle{
			ctx:     ctx,
			od:      od,
			Routing: Routing{router: NewOutputRouter()},
		}

		for _, opt := range opts {
			opt(lo)
		}

		od.ConfigureRoutine()
		return lo
	}

	return nil

}

func (lo *LiveOracle) EventLoop() error {

	oracleChannel := make(chan models.TransitData)

	go lo.od.ReadRoutine(lo.ctx, oracleChannel)

	for {
		select {
		case registerData := <-oracleChannel:
			lo.router.TransitOutput(registerData)

		case <-lo.ctx.Done():
			close(oracleChannel)

			return nil
		}

	}
}
