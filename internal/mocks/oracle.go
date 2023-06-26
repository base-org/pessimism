package mocks

import (
	"context"
	"math/big"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

type mockOracleDefinition struct {
}

func (md *mockOracleDefinition) ConfigureRoutine(core.PUUID) error {
	return nil
}

func (md *mockOracleDefinition) BackTestRoutine(_ context.Context, _ chan core.TransitData,
	_ *big.Int, _ *big.Int) error {
	return nil
}

func (md *mockOracleDefinition) ReadRoutine(_ context.Context, _ chan core.TransitData) error {
	return nil
}

// NewDummyOracle ... Takes in a register type that specifies the mocked output type
// Useful for testing inter-component connectivity and higher level component management abstractions
func NewDummyOracle(ctx context.Context, ot core.RegisterType, opts ...component.Option) (component.Component, error) {
	od := &mockOracleDefinition{}

	return component.NewOracle(ctx, ot, od, opts...)
}
