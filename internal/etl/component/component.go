package component

import (
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

// Component ... Generalized interface that all pipeline components must adhere to
type Component interface {
	AddEgress(core.ComponentID, chan core.TransitData) error
	RemoveEgress(core.ComponentID) error

	ID() core.ComponentID
	Type() core.ComponentType

	// EventLoop ... Component driver function; spun up as separate go routine
	EventLoop() error

	// GetIngress ... Returns component ingress channel for some register type value
	GetIngress(rt core.RegisterType) (chan core.TransitData, error)

	// OutputType ... Returns component output data type
	OutputType() core.RegisterType

	// TODO(#24): Add Internal Component Activity State Tracking
	ActivityState() ActivityState
}

// metaData ... Component-agnostic agnostic struct that stores component metadata and routing state
type metaData struct {
	id        core.ComponentID
	cType     core.ComponentType
	output    core.RegisterType
	state     ActivityState
	stateChan chan StateChange

	*ingressHandler
	*egressHandler

	*sync.RWMutex
}

func newMetaData(ct core.ComponentType, ot core.RegisterType) *metaData {
	return &metaData{
		id:             core.NilCompID(),
		cType:          ct,
		egressHandler:  newEgressHandler(),
		ingressHandler: newIngressHandler(),
		state:          Inactive,
		stateChan:      make(chan StateChange),
		output:         ot,
		RWMutex:        &sync.RWMutex{},
	}
}

// ActivityState ... Returns component current activity state
func (meta *metaData) ActivityState() ActivityState {
	return meta.state
}

// ID ... Returns component's ComponentID
func (meta *metaData) ID() core.ComponentID {
	return meta.id
}

// Type ... Returns component's type
func (meta *metaData) Type() core.ComponentType {
	return meta.cType
}

// OutputType ... Returns component's data output type
func (meta *metaData) OutputType() core.RegisterType {
	return meta.output
}

// emitStateChange ... Emits a stateChange event to stateChan
func (meta *metaData) emitStateChange(as ActivityState) {
	event := StateChange{
		ID:   meta.id,
		From: meta.state,
		To:   as,
	}

	meta.state = as
	meta.stateChan <- event
}

// TODO::Comment
type Option = func(*metaData)

func WithID(id core.ComponentID) Option {
	return func(meta *metaData) {
		meta.id = id
	}
}

func WithEventChan(sc chan StateChange) Option {
	return func(md *metaData) {
		md.stateChan = sc
	}
}
