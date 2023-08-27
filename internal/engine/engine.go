package engine

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"

	"go.uber.org/zap"
)

// Type ... Risk engine execution type
type Type int

const (
	HardCoded Type = iota + 1
	// NOTE: Dynamic heuristic support is not implemented
	Dynamic
)

// RiskEngine ... Execution engine interface
type RiskEngine interface {
	Type() Type
	Execute(context.Context, core.TransitData,
		heuristic.Heuristic) ([]*core.Activation, bool)
}

// hardCodedEngine ... Hard coded execution engine
// IE: native hardcoded application code for heuristic implementation
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

// Execute ... Executes the heuristic
func (e *hardCodedEngine) Execute(ctx context.Context, data core.TransitData,
	h heuristic.Heuristic) ([]*core.Activation, bool) {
	logger := logging.WithContext(ctx)

	logger.Debug("Performing heuristic activation",
		zap.String("suuid", h.SUUID().String()))
	outcome, activated, err := h.Assess(data)
	if err != nil {
		logger.Error("Failed to perform activation option for heuristic", zap.Error(err))

		metrics.WithContext(ctx).
			RecordAssessmentError(h)

		return nil, false
	}

	return outcome, activated
}
