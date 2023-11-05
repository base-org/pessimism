package process

import (
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

const (
	killSig = 0
)

type Process interface {
	/*
		NOTE - Storing the PathID assumes that one process
		can only be a part of one path at a time. This could be
		problematic if we want to have a process be a part of multiple
		paths at once. In that case, we would need to store a slice
		of PathIDs instead.
	*/
	Close() error
	EventLoop() error

	AddSubscriber(id core.ProcessID, outChan chan core.Event) error
	SetState(as ActivityState)

	AddRelay(tt core.TopicType) error
	GetRelay(tt core.TopicType) (chan core.Event, error)

	AddEngineRelay(relay *core.ExecInputRelay) error

	ID() core.ProcessID
	PathID() core.PathID

	Type() core.ProcessType
	EmitType() core.TopicType
	StateKey() *core.StateKey
	// TODO(#24): Add Internal Process Activity State Tracking
	ActivityState() ActivityState
}

// Process state
type State struct {
	id     core.ProcessID
	pathID core.PathID

	procType core.ProcessType
	publType core.TopicType
	as       ActivityState

	relay chan StateChange
	close chan int

	sk *core.StateKey

	*topics
	*subscribers

	*sync.RWMutex
}

func newState(pt core.ProcessType, tt core.TopicType) *State {
	return &State{
		id:     core.ProcessID{},
		pathID: core.PathID{},

		as:       Inactive,
		procType: pt,
		publType: tt,

		close: make(chan int),
		relay: make(chan StateChange),
		subscribers: &subscribers{
			subs: make(map[core.ProcIdentifier]chan core.Event),
		},
		topics: &topics{
			relays: make(map[core.TopicType]chan core.Event),
		},
		RWMutex: &sync.RWMutex{},
	}
}

func (s *State) ActivityState() ActivityState {
	return s.as
}

func (s *State) SetState(as ActivityState) {
	s.as = as
}

func (s *State) StateKey() *core.StateKey {
	return s.sk
}

func (s *State) ID() core.ProcessID {
	return s.id
}

func (s *State) PathID() core.PathID {
	return s.pathID
}

func (s *State) Type() core.ProcessType {
	return s.procType
}

func (s *State) EmitType() core.TopicType {
	return s.publType
}

func (s *State) emit(as ActivityState) {
	event := StateChange{
		ID:   s.id,
		From: s.as,
		To:   as,
	}

	s.as = as
	s.relay <- event // Send to upstream consumers
}

type Option = func(*State)

func WithID(id core.ProcessID) Option {
	return func(meta *State) {
		meta.id = id
	}
}
func WithPathID(id core.PathID) Option {
	return func(s *State) {
		s.pathID = id
	}
}

func WithEventChan(sc chan StateChange) Option {
	return func(s *State) {
		s.relay = sc
	}
}

func WithStateKey(sk *core.StateKey) Option {
	return func(s *State) {
		s.sk = sk
	}
}
