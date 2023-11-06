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
	Execute(ctx context.Context, data core.Event,
		h heuristic.Heuristic) (*heuristic.ActivationSet, error)
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
func (hce *hardCodedEngine) Execute(ctx context.Context, data core.Event,
	h heuristic.Heuristic) (*heuristic.ActivationSet, error) {
	logger := logging.WithContext(ctx)

	logger.Debug("Performing heuristic assessment",
		zap.String(logging.UUID, h.ID().ShortString()))
	as, err := h.Assess(data)
	if err != nil {
		logger.Error("Failed to perform activation option for heuristic", zap.Error(err),
			zap.String("heuristic_type", h.TopicType().String()))

		metrics.WithContext(ctx).
			RecordAssessmentError(h)

		return nil, err
	}

	return as, nil
}

// EventLoop ... Event loop for the risk engine
func (hce *hardCodedEngine) EventLoop(ctx context.Context) {
	logger := logging.WithContext(ctx)

	for {
		select {
		case <-ctx.Done(): // Context cancelled
			logger.Info("Risk engine event loop cancelled")
			return

		case args := <-hce.heuristicIn: // Heuristic input received
			logger.Debug("Heuristic input received",
				zap.String(logging.UUID, args.h.ID().ShortString()))

			start := time.Now()

			as, err := retry.Do[*heuristic.ActivationSet](ctx, 10, core.RetryStrategy(),
				func() (*heuristic.ActivationSet, error) {
					metrics.WithContext(ctx).RecordHeuristicRun(args.hi.PathID.Network(), args.h)
					return hce.Execute(ctx, args.hi.Input, args.h)
				})

			if err != nil {
				logger.Error("Failed to execute heuristic", zap.Error(err))
				metrics.WithContext(ctx).RecordAssessmentError(args.h)
			}

			metrics.WithContext(ctx).RecordAssessmentTime(args.h, float64(time.Since(start).Nanoseconds()))
			if as.Activated() {
				for _, act := range as.Entries() {
					alert := core.Alert{
						Timestamp:   act.TimeStamp,
						HeuristicID: args.h.ID(),
						HT:          args.h.Type(),
						Content:     act.Message,
						PathID:      args.hi.PathID,
						Net:         args.hi.PathID.Network(),
					}

					logger.Warn("Heuristic alert",
						zap.String(logging.UUID, args.h.ID().ShortString()),
						zap.String("heuristic_type", args.hi.PathID.String()),
						zap.String("message", act.Message))

					hce.alertEgress <- alert
				}
			}
		}
	}
}
