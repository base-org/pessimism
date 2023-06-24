package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/state"
)

// Analyzer ... Interface for analyzing pipelines
type Analyzer interface {
	Mergable(p1 Pipeline, p2 Pipeline) bool
	MergePipelines(ctx context.Context, p1 Pipeline, p2 Pipeline) (Pipeline, error)
}

// analyzer ... Implementation of Analyzer
type analyzer struct {
	dRegistry registry.Registry
}

// NewAnalyzer ... Initializer
func NewAnalyzer(dRegistry registry.Registry) Analyzer {
	return &analyzer{
		dRegistry: dRegistry,
	}
}

// Mergable ... Returns true if pipelines can be merged or deduped
func (a *analyzer) Mergable(p1 Pipeline, p2 Pipeline) bool {
	// Invalid if pipelines are not the same length
	if len(p1.Components()) != len(p2.Components()) {
		return false
	}

	// Invalid if pipelines are not live
	if p1.Config().PipelineType != core.Live ||
		p2.Config().PipelineType != core.Live {
		return false
	}

	// Invalid if either pipeline requires a backfill
	// NOTE - This is a temporary solution to prevent live backfills on two pipelines
	// from being merged.
	// In the future, this should only check the current state of each pipeline
	// to ensure that the backfill has been completed for both.
	if p1.Config().ClientConfig.Backfill() ||
		p2.Config().ClientConfig.Backfill() {
		return false
	}

	// Invalid if pipelines do not share the same PID
	if p1.UUID().PID != p2.UUID().PID {
		return false
	}

	return true
}

// MergePipelines ... Merges two pipelines into one (p1 --merge-> p2)
func (a *analyzer) MergePipelines(ctx context.Context, p1 Pipeline, p2 Pipeline) (Pipeline, error) {
	for i, compi := range p1.Components() {
		compj := p2.Components()[i]

		reg, err := a.dRegistry.GetRegister(compi.OutputType())
		if err != nil {
			return nil, err
		}

		if reg.Stateful() { // Merge state items from compi into compj
			err = a.mergeComponentState(ctx, compi, compj, p1.UUID(), p2.UUID())
			if err != nil {
				return nil, err
			}
		}
	}
	return p2, nil
}

// mergeComponentState ... Merges state items from p2 into p1
func (a *analyzer) mergeComponentState(ctx context.Context, compi, compj component.Component,
	p1, p2 core.PUUID) error {
	ss, err := state.FromContext(ctx)
	if err != nil {
		return err
	}

	items, err := ss.GetSlice(ctx, compi.StateKey())
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err := ss.SetSlice(ctx, compj.StateKey(), item)
		if err != nil {
			return err
		}
	}

	if compi.StateKey().IsNested() {
		err = a.MergeNestedStateKeys(ctx, compi, compj, p1, p2, ss)
		if err != nil {
			return err
		}
	}

	return nil
}

// MergeNestedStateKeys ... Merges nested state keys from p1 into p2
func (a *analyzer) MergeNestedStateKeys(ctx context.Context, c1, c2 component.Component,
	p1, p2 core.PUUID, ss state.Store) error {
	items, err := ss.GetSlice(ctx, c1.StateKey())
	if err != nil {
		return err
	}

	for _, item := range items {
		key1 := &core.StateKey{
			Prefix: c1.OutputType(),
			ID:     item,
			PUUID:  &p1,
		}

		key2 := &core.StateKey{
			Prefix: c2.OutputType(),
			ID:     item,
			PUUID:  &p2,
		}

		nestedValues, err := ss.GetSlice(ctx, key1)
		if err != nil {
			return err
		}

		for _, value := range nestedValues {
			_, err = ss.SetSlice(ctx, key2, value)
			if err != nil {
				return err
			}
		}

		err = ss.Remove(ctx, key1)
		if err != nil {
			return err
		}
	}

	return nil
}
