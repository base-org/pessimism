package component

import (
	"github.com/base-org/pessimism/internal/models"
)

type ActivityState = string

const (
	Inactive   = "inactive"
	Live       = "live"
	Terminated = "terminated"
)

// Component ... Generalized interface that all pipeline components must adhere to
type Component interface {
	// Routing functionality for downstream inter-component communication
	// Polymorphically extended from Router struct within
	AddDirective(models.ID, chan models.TransitData) error
	RemoveDirective(models.ID) error
	ID() models.ID
	Type() models.ComponentType

	// EventLoop ... Component driver function; spun up as separate go routine
	EventLoop() error

	// GetEntryPoint ... Returns component entrypoint channel for some register type value
	GetEntryPoint(rt models.RegisterType) (chan models.TransitData, error)

	CreateEntryPoint(rt models.RegisterType) error
	// OutputType ... Returns component output data type
	OutputType() models.RegisterType

	SetActivityState(s ActivityState)
	GetActivityState() ActivityState
}

// metaData ... Generalized component agnostic struct that stores component metadata and routing state
type metaData struct {
	id     models.ID
	cType  models.ComponentType
	output models.RegisterType
	state  ActivityState

	*ingress
	*router
}

// ID ... Returns
func (meta *metaData) SetActivityState(s ActivityState) {
	// NOTE - As of now anyone can set an arbitrary unverified state
	meta.state = s
}

// ID ... Returns
func (meta *metaData) GetActivityState() ActivityState {
	return meta.state
}

// ID ... Returns
func (meta *metaData) ID() models.ID {
	return meta.id
}

// Type ...
func (meta *metaData) Type() models.ComponentType {
	return meta.cType
}

// OutputType ...
func (meta *metaData) OutputType() models.RegisterType {
	return meta.output
}

type Option = func(*metaData)

func WithID(id models.ID) Option {
	return func(meta *metaData) {
		meta.id = id
	}
}
