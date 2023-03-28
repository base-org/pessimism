package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/conduit/models"
)

type OracleReadRoutine = func(chan models.TransitData)
type OracleConfigureRoutine = func(lo *LiveOracle) error

type OracleType = string

const (
	Backtest OracleType = "backtest"
	Live     OracleType = "live"
)

type ConnectionType = int

const (
	Poll   ConnectionType = 0
	Stream ConnectionType = 1
)

type Oracle interface {
	// TODO - Implement me
	Type() OracleType
}

type LiveOracleOption = func(*LiveOracle)

type LiveOracle struct {
	ct  ConnectionType
	ctx context.Context

	configureRoutine OracleConfigureRoutine
	readRoutine      OracleReadRoutine
	router           *OutputRouter

	InnerState map[string]interface{}
}

func (lo *LiveOracle) Type() OracleType {
	return Live
}

func NewLiveOracle(ctx context.Context, ot OracleType, orr OracleReadRoutine, ocr OracleConfigureRoutine, opts ...LiveOracleOption) Oracle {
	if ot == Live {
		lo := &LiveOracle{
			readRoutine:      orr,
			configureRoutine: ocr,
			router:           NewOutputRouter(),
		}

		for _, opt := range opts {
			opt(lo)
		}

		ocr(lo)
		return lo
	}

	return nil

}

func (lo *LiveOracle) EventLoop() error {

	oracleChannel := make(chan models.TransitData)

	for {
		select {
		case registerData := <-oracleChannel:
			lo.router.TransitOutput(registerData)

		case <-lo.ctx.Done():
			return nil
		}

	}
}
