package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/state"
)

type Analyzer interface {
	Mergable(p1 Pipeline, p2 Pipeline) bool
	MergePipelines(ctx context.Context, p1 Pipeline, p2 Pipeline) (Pipeline, error)
}

type analyzer struct {
	dRegistry registry.Registry
}

func NewAnalyzer(dRegistry registry.Registry) Analyzer {
	return &analyzer{
		dRegistry: dRegistry,
	}
}

func (a *analyzer) Mergable(p1 Pipeline, p2 Pipeline) bool {
	if len(p1.Components()) != len(p2.Components()) {
		return false
	}

	if p1.Config().PipelineType != core.Live ||
		p2.Config().PipelineType != core.Live {
		return false
	}

	if p1.Config().ClientConfig.Backfill() ||
		p2.Config().ClientConfig.Backfill() {
		return false
	}

	if p1.UUID().PID != p2.UUID().PID {
		return false
	}

	return p1.UUID() != p2.UUID()

}

func (a *analyzer) MergePipelines(ctx context.Context, p1 Pipeline, p2 Pipeline) (Pipeline, error) {
	for i, compi := range p1.Components() {
		compj := p2.Components()[i]

		reg, err := a.dRegistry.GetRegister(compi.OutputType())
		if err != nil {
			return nil, err
		}

		if reg.Stateful() { // Merge state items from p2 into p1
			ss, err := state.FromContext(ctx)
			if err != nil {
				return nil, err
			}

			sliceItems, err := ss.GetSlice(ctx, compi.StateKey())
			if err != nil {
				return nil, err
			}

			for _, item := range sliceItems {
				ss.SetSlice(ctx, compj.StateKey(), item)
			}

			if compi.StateKey().IsNested() {

				for _, item := range sliceItems {
					nestedKey := state.MakeKey(core.NestedPrefix, item, false).WithPUUID(p1.UUID())
					nestedKeyj := state.MakeKey(core.NestedPrefix, item, false).WithPUUID(p2.UUID())

					nestedValues, err := ss.GetSlice(ctx, nestedKey)

					for _, value := range nestedValues {
						_, err = ss.SetSlice(ctx, nestedKeyj, value)
						if err != nil {
							return nil, err
						}
					}

					err = ss.Remove(ctx, nestedKey)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}
	return p2, nil

}
