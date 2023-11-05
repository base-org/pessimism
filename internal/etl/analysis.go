package etl

import (
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
)

// Analyzer ... Interface for analyzing pipelines
type Analyzer interface {
	Mergable(p1 Path, p2 Path) bool
	// MergePipelines(ctx context.Context, p1 Pipeline, p2 Pipeline) (Pipeline, error)
}

// analyzer ... Implementation of Analyzer
type analyzer struct {
	r *registry.Registry
}

// NewAnalyzer ... Initializer
func NewAnalyzer(r *registry.Registry) Analyzer {
	return &analyzer{
		r: r,
	}
}

// Mergable ... Returns true if pipelines can be merged or deduped
func (a *analyzer) Mergable(path1 Path, path2 Path) bool {
	// Invalid if pipelines are not the same length
	if len(path1.Processes()) != len(path2.Processes()) {
		return false
	}

	// Invalid if pipelines are not live
	if path1.Config().PathType != core.Live ||
		path2.Config().PathType != core.Live {
		return false
	}

	// Invalid if either pipeline requires a backfill
	// NOTE - This is a temporary solution to prevent live backfills on two pipelines
	// from being merged.
	// In the future, this should only check the current state of each pipeline
	// to ensure that the backfill has been completed for both.
	if path1.Config().ClientConfig.Backfill() ||
		path2.Config().ClientConfig.Backfill() {
		return false
	}

	// Invalid if pipelines do not share the same PID
	if path1.UUID().ID != path2.UUID().ID {
		return false
	}

	return true
}

// NOTE - This is intentionally commented out for now as its not in-use.

// // MergePipelines ... Merges two pipelines into one (p1 --merge-> p2)
// func (a *analyzer) MergePipelines(ctx context.Context, p1 Pipeline, p2 Pipeline) (Pipeline, error) {
// 	for i, compi := range p1.Processs() {
// 		compj := p2.Processs()[i]

// 		reg, err := a.r.GetDataTopic(compi.OutputType())
// 		if err != nil {
// 			return nil, err
// 		}

// 		if reg.Stateful() { // Merge state items from compi into compj
// 			err = a.mergeProcessState(ctx, compi, compj, p1.UUID(), p2.UUID())
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 	}
// 	return p2, nil
// }

// // mergeProcessState ... Merges state items from p2 into p1
// func (a *analyzer) mergeProcessState(ctx context.Context, compi, compj processProcess,
// 	p1, p2 core.PathID) error {
// 	ss, err := state.FromContext(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	items, err := ss.GetSlice(ctx, compi.StateKey())
// 	if err != nil {
// 		return err
// 	}

// 	for _, item := range items {
// 		_, err := ss.SetSlice(ctx, compj.StateKey(), item)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	if compi.StateKey().IsNested() {
// 		err = a.MergeNestedStateKeys(ctx, compi, compj, p1, p2, ss)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// // MergeNestedStateKeys ... Merges nested state keys from p1 into p2
// func (a *analyzer) MergeNestedStateKeys(ctx context.Context, c1, c2 processProcess,
// 	p1, p2 core.PathID, ss state.Store) error {
// 	items, err := ss.GetSlice(ctx, c1.StateKey())
// 	if err != nil {
// 		return err
// 	}

// 	for _, item := range items {
// 		key1 := &core.StateKey{
// 			Prefix: c1.OutputType(),
// 			ID:     item,
// 			PathID:  &p1,
// 		}

// 		key2 := &core.StateKey{
// 			Prefix: c2.OutputType(),
// 			ID:     item,
// 			PathID:  &p2,
// 		}

// 		nestedValues, err := ss.GetSlice(ctx, key1)
// 		if err != nil {
// 			return err
// 		}

// 		for _, value := range nestedValues {
// 			_, err = ss.SetSlice(ctx, key2, value)
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		err = ss.Remove(ctx, key1)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
