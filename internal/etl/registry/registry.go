package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

const (
	noEntryErr = "could not find data topic type %s"
)

type Registry struct {
	topics map[core.TopicType]*core.DataTopic
}

func New() *Registry {
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

	return &Registry{topics}
}

func makeDeps(types ...core.TopicType) []core.TopicType {
	deps := make([]core.TopicType, len(types))
	copy(deps, types)

	return deps
}

func noDeps() []core.TopicType {
	return []core.TopicType{}
}

func noState() *core.StateKey {
	return nil
}

// Returns in-order slice of ETL path path
func (r *Registry) TopicPath(tt core.TopicType) (core.TopicPath, error) {
	topic, err := r.GetDataTopic(tt)
	if err != nil {
		return core.TopicPath{}, err
	}

	topics := make([]*core.DataTopic, len(topic.Dependencies)+1)

	topics[0] = topic

	for i, depType := range topic.Dependencies {
		depRegister, err := r.GetDataTopic(depType)
		if err != nil {
			return core.TopicPath{}, err
		}

		topics[i+1] = depRegister
	}

	return core.TopicPath{Path: topics}, nil
}

func (r *Registry) GetDataTopic(tt core.TopicType) (*core.DataTopic, error) {
	if _, exists := r.topics[tt]; !exists {
		return nil, fmt.Errorf(noEntryErr, tt)
	}

	return r.topics[tt], nil
}
