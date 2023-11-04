package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

const (
	noEntryErr = "could not find entry in registry for encoded register type %s"
)

type Registry struct {
	topics map[core.TopicType]*core.DataTopic
}

func NewRegistry() Registry {
	topics := map[core.TopicType]*core.DataTopic{
		core.BlockHeader: {
			Addressing:  false,
			DataType:    core.BlockHeader,
			ProcessType: core.Read,
			Constructor: NewHeaderTraversal,

			Dependencies: noDeps(),
			Sk:           noState(),
		},

		core.Log: {
			Addressing:  true,
			DataType:    core.Log,
			ProcessType: core.Subscribe,
			Constructor: NewLogSubscriber,

			Dependencies: makeDeps(core.BlockHeader),
			Sk: &core.StateKey{
				Nesting: true,
				Prefix:  core.Log,
				ID:      core.AddressKey,
				PathID:  nil,
			},
		},
	}

	return Registry{topics}
}

// makeDeps ... Makes dependency slice
func makeDeps(types ...core.TopicType) []core.TopicType {
	deps := make([]core.TopicType, len(types))
	copy(deps, types)

	return deps
}

// noDeps ... Returns empty dependency slice
func noDeps() []core.TopicType {
	return []core.TopicType{}
}

// noState ... Returns empty state key, indicating no state dependencies
// for cross subsystem communication (i.e. ETL -> Risk Engine)
func noState() *core.StateKey {
	return nil
}

// GetDependencyPath ... Returns in-order slice of ETL pipeline path
func (cr *Registry) GetDependencyPath(rt core.TopicType) (core.RegisterDependencyPath, error) {
	destRegister, err := cr.GetDataTopic(rt)
	if err != nil {
		return core.RegisterDependencyPath{}, err
	}

	topics := make([]*core.DataTopic, len(destRegister.Dependencies)+1)

	topics[0] = destRegister

	for i, depType := range destRegister.Dependencies {
		depRegister, err := cr.GetDataTopic(depType)
		if err != nil {
			return core.RegisterDependencyPath{}, err
		}

		topics[i+1] = depRegister
	}

	return core.RegisterDependencyPath{Path: topics}, nil
}

// GetDataTopic ... Returns a data register provided an enum type
func (cr *Registry) GetDataTopic(rt core.TopicType) (*core.DataTopic, error) {
	if _, exists := cr.topics[rt]; !exists {
		return nil, fmt.Errorf(noEntryErr, rt)
	}

	return cr.topics[rt], nil
}
