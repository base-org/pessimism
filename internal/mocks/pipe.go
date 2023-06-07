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
func (md *mockPipeDefinition) ConfigureRoutine(core.PipelineUUID) error {
	return nil
}

// Transform ... Mocked transform function that returns an empty slice
func (md *mockPipeDefinition) Transform(ctx context.Context, data core.TransitData) ([]core.TransitData, error) {
	return []core.TransitData{}, nil
}

// NewMockPipe ... Takes in a register type that specifies the mocked output type
// Useful for testing inter-component connectivity and higher level component management abstractions
func NewMockPipe(ctx context.Context, ot core.RegisterType) (component.Component, error) {
	od := &mockPipeDefinition{}

	return component.NewPipe(ctx, od, core.BlackholeTX, ot)
}
