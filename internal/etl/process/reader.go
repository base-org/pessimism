package process

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

// Routine ...
type Routine interface {
	Loop(ctx context.Context, componentChan chan core.Event) error
	Height() (*big.Int, error)
}

// ChainReader ...
type ChainReader struct {
	ctx context.Context

	routine   Routine
	jobEvents chan core.Event

	wg *sync.WaitGroup

	*State
}

// NewReader ... Initializer
func NewReader(ctx context.Context, outType core.TopicType,
	r Routine, opts ...Option) (Process, error) {
	cr := &ChainReader{
		ctx:       ctx,
		routine:   r,
		jobEvents: make(chan core.Event),
		wg:        &sync.WaitGroup{},
		State:     newState(core.Read, outType),
	}

	for _, opt := range opts {
		opt(cr.State)
	}

	logging.WithContext(ctx).Info("Constructed process",
		zap.String(logging.Process, cr.State.id.String()))

	return cr, nil
}

func (cr *ChainReader) Height() (*big.Int, error) {
	return cr.routine.Height()
}

func (cr *ChainReader) Close() error {
	cr.wg.Wait()
	return nil
}

// EventLoop ...
func (cr *ChainReader) EventLoop() error {
	// TODO(#24) - Add Internal Component Activity State Tracking

	logger := logging.WithContext(cr.ctx)

	logger.Debug("Starting process job",
		zap.String(logging.Process, cr.id.String()))

	cr.wg.Add(1)

	jobCtx, cancel := context.WithCancel(cr.ctx)

	// Run job
	go func() {
		defer cr.wg.Done()
		if err := cr.routine.Loop(jobCtx, cr.jobEvents); err != nil {
			logger.Error("Received error from read routine",
				zap.String(logging.Process, cr.id.String()),
				zap.Error(err))
		}
	}()

	for {
		select {
		case event := <-cr.jobEvents:
			logger.Debug("Sending event to subscribers",
				zap.String(logging.Process, cr.id.String()),
				zap.String("event", event.Type.String()))

			if err := cr.subscribers.Publish(event); err != nil {
				logger.Error(relayErr, zap.String("ID", cr.id.String()))
			}

			if cr.subscribers.None() {
				latency := float64(time.Since(event.OriginTS).Milliseconds())
				metrics.WithContext(cr.ctx).
					RecordPathLatency(cr.PathID(), latency)
			}

		case <-cr.close:
			logger.Debug("Shutting down process",
				zap.String(logging.Process, cr.id.String()))
			close(cr.jobEvents)
			cancel()
			return nil
		}
	}
}
