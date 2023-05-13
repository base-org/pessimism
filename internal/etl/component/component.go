package component

import (
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

const (
	killSig = 0
)

// Component ... Generalized interface that all pipeline components must adhere to
type Component interface {
	// UUID ...
	UUID() core.ComponentUUID
	// Type ... Returns component enum type
	Type() core.ComponentType

	// AddRelay ... Adds an engine relay to component egress routing
	AddRelay(relay *core.EngineInputRelay) error

	// AddEgress ...
	AddEgress(core.ComponentUUID, chan core.TransitData) error
	// RemoveEgress ...
	RemoveEgress(core.ComponentUUID) error

	// Close ... Signifies a component to stop operating
	Close() error

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
	id     core.ComponentUUID
	cType  core.ComponentType
	output core.RegisterType
	state  ActivityState

	closeChan chan int
	stateChan chan StateChange

	*ingressHandler
	*egressHandler

	*sync.RWMutex
}

// newMetaData ... Initializer
func newMetaData(ct core.ComponentType, ot core.RegisterType) *metaData {
	return &metaData{
		id:             core.NilComponentUUID(),
		cType:          ct,
		egressHandler:  newEgressHandler(),
		ingressHandler: newIngressHandler(),
		state:          Inactive,
		closeChan:      make(chan int),
		stateChan:      make(chan StateChange),
		output:         ot,
		RWMutex:        &sync.RWMutex{},
	}
}

// ActivityState ... Returns component current activity state
func (meta *metaData) ActivityState() ActivityState {
	return meta.state
}

// UUID ... Returns component's ComponentUUID
func (meta *metaData) UUID() core.ComponentUUID {
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
	meta.stateChan <- event // Send to upstream consumers
}

// Option ... Component type agnostic option
type Option = func(*metaData)

// WithID ... Passes component UUID to component metadata field
func WithID(id core.ComponentUUID) Option {
	return func(meta *metaData) {
		meta.id = id
	}
}

// WithEventChan ... Passes state channel to component metadata field
func WithEventChan(sc chan StateChange) Option {
	return func(md *metaData) {
		md.stateChan = sc
	}
}
