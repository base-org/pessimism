package registry

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
)

type mockGethBlockDefinition struct {
}

func (md *mockGethBlockDefinition) ConfigureRoutine() error {
	return nil
}
func (md *mockGethBlockDefinition) BackTestRoutine(ctx context.Context, componentChan chan core.TransitData) error {
	return nil
}
func (md *mockGethBlockDefinition) ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error {
	return nil
}

func NewMockOracle(ctx context.Context, ot core.RegisterType) (component.Component, error) {
	od := &mockGethBlockDefinition{}

	return component.NewOracle(ctx, 0, ot, od)
}
