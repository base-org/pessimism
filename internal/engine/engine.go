package engine

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

type EngineType int

const (
	HardCoded EngineType = iota
)

type RiskEngine interface {
	Type() EngineType
	EventLoop(ctx context.Context) error
}

type engine struct {
	is       *InvariantStore
	readChan chan core.TransitData
}

func NewEngine(readChan chan core.TransitData, is *InvariantStore) RiskEngine {
	return &engine{
		is:       is,
		readChan: readChan,
	}
}

func (e engine) Type() EngineType {
	return HardCoded
}

func (e engine) EventLoop(ctx context.Context) error {
	logger := logging.WithContext(ctx)

	for {
		select {
		case data := <-e.readChan:

			invs, err := e.is.GetInvariants(data.Type)
			if err != nil {
				logger.Error("Could not get invariants from store")
			}

			for _, inv := range invs {
				invalid, err := inv.Invalidate(data)
				if err != nil {
					logger.Error(err.Error())
				}

				if invalid {
					logger.Info("Invariant invalidatio occured")
				}
			}

		case <-ctx.Done():
			logger.Info("Received shutdown for risk engine event loop")
			return nil

		}

	}

}
