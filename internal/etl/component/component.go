package component

import (
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

// Component ... Generalized interface that all pipeline components must adhere to
type Component interface {
	// Routing functionality for downstream inter-component communication
	// Polymorphically extended from Router struct within
	AddDirective(core.ComponentID, chan core.TransitData) error
	RemoveDirective(core.ComponentID) error
	ID() core.ComponentID
	Type() core.ComponentType

	// EventLoop ... Component driver function; spun up as separate go routine
	EventLoop() error

	// GetEntryPoint ... Returns component entrypoint channel for some register type value
	GetEntryPoint(rt core.RegisterType) (chan core.TransitData, error)

	// OutputType ... Returns component output data type
	OutputType() core.RegisterType

	GetActivityState() ActivityState
}

// metaData ... Generalized component agnostic struct that stores component metadata and routing state
type metaData struct {
	id     core.ComponentID
	cType  core.ComponentType
	output core.RegisterType
	state  ActivityState

	*ingress
	*router
	*sync.RWMutex
}

// ID ... Returns
func (meta *metaData) GetActivityState() ActivityState {
	return meta.state
}

// ID ... Returns
func (meta *metaData) ID() core.ComponentID {
	return meta.id
}

// Type ...
func (meta *metaData) Type() core.ComponentType {
	return meta.cType
}

// OutputType ...
func (meta *metaData) OutputType() core.RegisterType {
	return meta.output
}

type Option = func(*metaData)

func WithID(id core.ComponentID) Option {
	return func(meta *metaData) {
		meta.id = id
	}
}
