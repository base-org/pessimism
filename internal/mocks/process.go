package mocks

import (
	"context"
	big "math/big"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/process"
)

type mockSubscription struct {
}

func (ms *mockSubscription) Run(_ context.Context, e core.Event) ([]core.Event, error) {
	return []core.Event{e}, nil
}

func NewSubscriber(ctx context.Context, it core.TopicType, ot core.TopicType,
	opts ...process.Option) (process.Process, error) {
	ms := &mockSubscription{}

	return process.NewSubscriber(ctx, ms, it, ot, opts...)
}

type mockTraversal struct {
}

func (md *mockTraversal) ConfigureRoutine(core.PathID) error {
	return nil
}

func (md *mockTraversal) Loop(_ context.Context, _ chan core.Event) error {
	return nil
}

func (md *mockTraversal) Height() (*big.Int, error) {
	return big.NewInt(0), nil
}

// NewReader
func NewReader(ctx context.Context, ot core.TopicType, opts ...process.Option) (process.Process, error) {
	mt := &mockTraversal{}

	return process.NewReader(ctx, ot, mt, opts...)
}
