package engine

import (
	"context"
	"time"

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

// ExecInput ... Parameter wrapper for engine execution input
type ExecInput struct {
	ctx context.Context
	hi  core.HeuristicInput
	h   heuristic.Heuristic
}

// RiskEngine ... Execution engine interface
type RiskEngine interface {
	Type() Type
	Execute(context.Context, core.TransitData,
		heuristic.Heuristic) (*core.Activation, bool)
	AddWorkerIngress(chan ExecInput)
	EventLoop(context.Context)
}

// hardCodedEngine ... Hard coded execution engine
// IE: native hardcoded application code for heuristic implementation
type hardCodedEngine struct {
	heuristicIn chan ExecInput
	alertEgress chan core.Alert
}

// NewHardCodedEngine ... Initializer
func NewHardCodedEngine(egress chan core.Alert) RiskEngine {
	return &hardCodedEngine{
		alertEgress: egress,
	}
}

// Type ... Returns the engine type
func (hce *hardCodedEngine) Type() Type {
	return HardCoded
}

// AddWorkerIngress ... Adds a worker ingress channel
func (hce *hardCodedEngine) AddWorkerIngress(ingress chan ExecInput) {
	hce.heuristicIn = ingress
}

// Execute ... Executes the heuristic
func (hce *hardCodedEngine) Execute(ctx context.Context, data core.TransitData,
	h heuristic.Heuristic) (*core.Activation, bool) {
	logger := logging.WithContext(ctx)

	logger.Debug("Performing heuristic activation",
		zap.String(logging.SUUIDKey, h.SUUID().String()))
	outcome, activated, err := h.Assess(data)
	if err != nil {
		logger.Error("Failed to perform activation option for heuristic", zap.Error(err))

		metrics.WithContext(ctx).
			RecordAssessmentError(h)

		return nil, false
	}

	return outcome, activated
}

// EventLoop ... Event loop for the risk engine
func (hce *hardCodedEngine) EventLoop(ctx context.Context) {
	logger := logging.WithContext(ctx)

	for {
		select {
		case <-ctx.Done(): // Context cancelled
			logger.Info("Risk engine event loop cancelled")
			return

		case execInput := <-hce.heuristicIn: // Heuristic input received
			logger.Debug("Heuristic input received",
				zap.String(logging.SUUIDKey, execInput.h.SUUID().String()))

			// (1) Execute heuristic
			start := time.Now()
			outcome, activated := hce.Execute(ctx, execInput.hi.Input, execInput.h)
			metrics.WithContext(ctx).RecordHeuristicRun(execInput.h)
			metrics.WithContext(ctx).RecordInvExecutionTime(execInput.h, float64(time.Since(start).Nanoseconds()))

			// (2) Send alert if activated
			if activated {
				alert := core.Alert{
					Timestamp: outcome.TimeStamp,
					SUUID:     execInput.h.SUUID(),
					Content:   outcome.Message,
					PUUID:     execInput.hi.PUUID,
					Ptype:     execInput.hi.PUUID.PipelineType(),
				}

				logger.Warn("Heuristic alert",
					zap.String(logging.SUUIDKey, execInput.h.SUUID().String()),
					zap.String("message", outcome.Message))

				hce.alertEgress <- alert
			}
		}
	}
}
