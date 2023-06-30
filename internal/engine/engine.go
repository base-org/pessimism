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
	Dynamic
)

// RiskEngine ... Execution engine interface
type RiskEngine interface {
	Type() Type
	Execute(context.Context, core.TransitData,
		invariant.Invariant) (*core.InvalOutcome, bool)
}

// hardCodedEngine ... Hard coded execution engine
// IE: native application code for invariant implementation
type hardCodedEngine struct {
	// TODO: Add any engine specific fields here
}

// NewHardCodedEngine ... Initializer
func NewHardCodedEngine() RiskEngine {
	return &hardCodedEngine{}
}

// Type ... Returns the engine type
func (e *hardCodedEngine) Type() Type {
	return HardCoded
}

// Execute ... Executes the invariant
func (e *hardCodedEngine) Execute(ctx context.Context, data core.TransitData,
	inv invariant.Invariant) (*core.InvalOutcome, bool) {
	logger := logging.WithContext(ctx)

	logger.Debug("Performing invariant invalidation",
		zap.String("suuid", inv.SUUID().String()))
	outcome, invalid, err := inv.Invalidate(data)
	if err != nil {
		logger.Error("Failed to perform invalidation option for invariant", zap.Error(err))
		return nil, false
	}

	return outcome, invalid
}
