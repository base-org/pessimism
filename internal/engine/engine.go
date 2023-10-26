package engine

import (
	"context"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/ethereum-optimism/optimism/op-service/retry"

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
		heuristic.Heuristic) *heuristic.ActivationSet
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
	h heuristic.Heuristic) *heuristic.ActivationSet {
	logger := logging.WithContext(ctx)

	logger.Debug("Performing heuristic activation",
		zap.String(logging.SUUIDKey, h.SUUID().String()))
	activationSet, err := h.Assess(data)
	if err != nil {
		logger.Error("Failed to perform activation option for heuristic", zap.Error(err))

		metrics.WithContext(ctx).
			RecordAssessmentError(h)

		return nil
	}

	return activationSet
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

			// (1) Execute heuristic with retry strategy
			start := time.Now()

			var actSet *heuristic.ActivationSet

			retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
			if _, err := retry.Do[interface{}](ctx, 10, retryStrategy, func() (interface{}, error) {
				actSet = hce.Execute(ctx, execInput.hi.Input, execInput.h)
				metrics.WithContext(ctx).RecordHeuristicRun(execInput.h)
				metrics.WithContext(ctx).RecordInvExecutionTime(execInput.h, float64(time.Since(start).Nanoseconds()))
				// a-ok!
				return nil, nil
			}); err != nil {
				logger.Error("Failed to execute heuristic", zap.Error(err))
				metrics.WithContext(ctx).RecordAssessmentError(execInput.h)
			}

			// (2) Send alerts for respective activations
			if actSet.Activated() {
				for _, act := range actSet.Entries() {
					alert := core.Alert{
						Timestamp: act.TimeStamp,
						SUUID:     execInput.h.SUUID(),
						Content:   act.Message,
						PUUID:     execInput.hi.PUUID,
						Ptype:     execInput.hi.PUUID.PipelineType(),
					}

					logger.Warn("Heuristic alert",
						zap.String(logging.SUUIDKey, execInput.h.SUUID().String()),
						zap.String("message", act.Message))

					hce.alertEgress <- alert
				}
			}
		}
	}
}
