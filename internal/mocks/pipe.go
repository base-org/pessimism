package mocks

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

// mockPipeDefinition ... Mocked pipe definition struct
type mockPipeDefinition struct {
}

// ConfigureRoutine ... Mocked configure routine function that returns nil
func (md *mockPipeDefinition) ConfigureRoutine(core.PUUID) error {
	return nil
}

// Transform ... Mocked transform function that returns an empty slice
func (md *mockPipeDefinition) Transform(_ context.Context, td core.TransitData) ([]core.TransitData, error) {
	return []core.TransitData{td}, nil
}

// NewDummyPipe ... Takes in a register type that specifies the mocked output type
// Useful for testing inter-component connectivity and higher level component management abstractions
func NewDummyPipe(ctx context.Context, it core.RegisterType, ot core.RegisterType,
	opts ...component.Option) (component.Component, error) {
	pd := &mockPipeDefinition{}

	return component.NewPipe(ctx, pd, it, ot, opts...)
}
