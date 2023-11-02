package mocks

import (
	"context"
	"math/big"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

type mockTraversal struct {
}

func (md *mockTraversal) ConfigureRoutine(core.PUUID) error {
	return nil
}

func (md *mockTraversal) Loop(_ context.Context, _ chan core.TransitData) error {
	return nil
}

func (md *mockTraversal) Height() (*big.Int, error) {
	return big.NewInt(0), nil
}

// NewDummyReader
func NewDummyReader(ctx context.Context, ot core.RegisterType, opts ...component.Option) (component.Component, error) {
	mt := &mockTraversal{}

	return component.NewReader(ctx, ot, mt, opts...)
}
