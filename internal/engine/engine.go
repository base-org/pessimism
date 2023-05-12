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
	Execute(ctx context.Context, data core.TransitData, inv invariant.Invariant) error
}

type hardCodedEngine struct {
}

func NewHardCodedEngine() RiskEngine {
	return &hardCodedEngine{}
}

func (e *hardCodedEngine) Type() Type {
	return HardCoded
}

func (e *hardCodedEngine) Execute(ctx context.Context, data core.TransitData, inv invariant.Invariant) error {
	logger := logging.WithContext(ctx)

	logger.Debug("Performing invariant invalidation",
		zap.String("suuid", inv.UUID().String()))
	invalid, err := inv.Invalidate(data)
	if err != nil {
		logger.Error("Failed to perform invalidation option for invariant", zap.Error(err))
		return err
	}

	if invalid {
		logger.Info("Invariant invalidation occurred", zap.String("suuid", inv.UUID().String()))
	}

	return nil
}
