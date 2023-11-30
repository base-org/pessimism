//go:generate mockgen -package mocks --destination ../../mocks/heuristic.go . Heuristic

package heuristic

import (
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
)

// ExecutionType ... Enum for execution type
type ExecutionType int

const (
	// HardCoded ... Hard coded execution type (ie native application code)
	HardCoded ExecutionType = iota

	invalidTopicErr = "invalid input type provided for heuristic. expected %s, got %s"
)

type Heuristic interface {
	TopicType() core.TopicType
	Validate(core.Event) error
	Assess(e core.Event) (*ActivationSet, error)
	Type() core.HeuristicType
	ID() core.UUID
	SetID(core.UUID)
}

type BaseHeuristicOpt = func(bh *BaseHeuristic) *BaseHeuristic

type BaseHeuristic struct {
	ht    core.HeuristicType
	id    core.UUID
	topic core.TopicType
}

func New(topic core.TopicType, t core.HeuristicType,
	opts ...BaseHeuristicOpt) Heuristic {
	bi := &BaseHeuristic{
		ht:    t,
		topic: topic,
	}

	for _, opt := range opts {
		opt(bi)
	}

	return bi
}

func (bi *BaseHeuristic) Type() core.HeuristicType {
	return bi.ht
}

func (bi *BaseHeuristic) ID() core.UUID {
	return bi.id
}

func (bi *BaseHeuristic) TopicType() core.TopicType {
	return bi.topic
}

func (bi *BaseHeuristic) Assess(_ core.Event) (*ActivationSet, error) {
	return NoActivations(), nil
}

func (bi *BaseHeuristic) SetID(id core.UUID) {
	bi.id = id
}

func (bi *BaseHeuristic) Validate(e core.Event) error {
	if e.Type != bi.TopicType() {
		return fmt.Errorf(invalidTopicErr, bi.TopicType(), e.Type)
	}

	return nil
}

type Activation struct {
	TimeStamp time.Time
	Message   string
}

type ActivationSet struct {
	acts []*Activation
}

func NewActivationSet() *ActivationSet {
	return &ActivationSet{
		acts: make([]*Activation, 0),
	}
}

func (as *ActivationSet) Len() int {
	return len(as.acts)
}

func (as *ActivationSet) Add(a *Activation) *ActivationSet {
	as.acts = append(as.acts, a)
	return as
}

func (as *ActivationSet) Entries() []*Activation {
	return as.acts
}

func (as *ActivationSet) Activated() bool {
	return as.Len() > 0
}

func NoActivations() *ActivationSet {
	return NewActivationSet()
}
