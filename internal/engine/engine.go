package engine

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

type Type int

const (
	HardCoded Type = iota
)

type RiskEngine interface {
	Type() Type
	Execute(context.Context, core.TransitData, invariant.Invariant) (*core.Alert, error)
}

type hardCodedEngine struct {
}

func NewHardCodedEngine() RiskEngine {
	return &hardCodedEngine{}
}

func (e *hardCodedEngine) Type() Type {
	return HardCoded
}

func (e *hardCodedEngine) Execute(ctx context.Context, data core.TransitData,
	inv invariant.Invariant) (*core.Alert, error) {
	logger := logging.WithContext(ctx)

	logger.Debug("Performing invariant invalidation",
		zap.String("suuid", inv.UUID().String()))
	outcome, err := inv.Invalidate(data)
	if err != nil {
		logger.Error("Failed to perform invalidation option for invariant", zap.Error(err))
		return nil, err
	}

	if outcome == nil {
		return nil, nil
	}

	alert := core.NewAlert(outcome.TimeStamp, outcome.SUUID, outcome.Message)

	return &alert, nil
}
