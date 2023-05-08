package engine

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
)

type Type int

const (
	HardCoded Type = iota
)

type RiskEngine interface {
	Type() Type
	Execute(ctx context.Context, data core.TransitData, invs []invariant.Invariant) error
}

type hardCodedEngine struct {
}

func NewHardCodedEngine() RiskEngine {
	return &hardCodedEngine{}
}

func (e hardCodedEngine) Type() Type {
	return HardCoded
}

func (e hardCodedEngine) Execute(ctx context.Context, data core.TransitData, invs []invariant.Invariant) error {
	logger := logging.WithContext(ctx)

	for _, inv := range invs {
		invalid, err := inv.Invalidate(data)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		if invalid {
			logger.Info("Invariant invalidation occurred")
		}
	}

	return nil
}
