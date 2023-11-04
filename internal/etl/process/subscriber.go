package process

import (
	"context"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

type Subscription interface {
	Run(ctx context.Context, data core.Event) ([]core.Event, error)
}

type Subscriber struct {
	ctx context.Context
	tt  core.TopicType

	spt Subscription

	*State
}

// NewSubscriber ... Initializer
func NewSubscriber(ctx context.Context, s Subscription, tt core.TopicType,
	outType core.TopicType, opts ...Option) (Process, error) {

	sub := &Subscriber{
		ctx: ctx,
		spt: s,
		tt:  tt,

		State: newState(core.Subscribe, outType),
	}

	if err := sub.AddRelay(tt); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(sub.State)
	}

	return sub, nil
}

func (sub *Subscriber) Close() error {
	sub.close <- killSig

	return nil
}

func (sub *Subscriber) EventLoop() error {
	logger := logging.WithContext(sub.ctx)

	logger.Info("Starting event loop",
		zap.String("ID", sub.id.String()),
	)

	relay, err := sub.GetRelay(sub.tt)
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-relay:

			events, err := sub.spt.Run(sub.ctx, event)
			if err != nil {
				logger.Error(err.Error(), zap.String("ID", sub.id.String()))
			}

			if sub.subscribers.None() {
				latency := float64(time.Since(event.OriginTS).Milliseconds())

				metrics.WithContext(sub.ctx).
					RecordPathLatency(sub.pathID, latency)
			}

			length := len(events)
			logger.Debug("Received publisher events",
				zap.String(logging.Process, sub.id.String()),
				zap.Int("Length", length))

			if length == 0 {
				continue
			}

			logger.Debug("Sending data batch",
				zap.String("ID", sub.id.String()),
				zap.String("Type", sub.EmitType().String()))

			if err := sub.subscribers.PublishBatch(events); err != nil {
				logger.Error(relayErr, zap.String("ID", sub.id.String()))
			}

		// Manager is telling us to shutdown
		case <-sub.close:
			logger.Debug("Received component shutdown signal",
				zap.String("ID", sub.id.String()))

			// p.emitStateChange(Terminated)

			return nil
		}
	}
}
