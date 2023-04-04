package pipeline

import (
	"github.com/base-org/pessimism/internal/models"
)

// Component ... Generalized interface that all pipeline components must adhere to
type Component interface {
	// Routing functionality for downstream inter-component communication
	// Polymorphically extended from Router struct within
	AddDirective(models.ComponentID, chan models.TransitData) error
	RemoveDirective(models.ComponentID) error
	ID() models.ComponentID
	Type() models.ComponentType

	// EventLoop ... Component driver function; spun up as separate go routine
	EventLoop() error

	// EntryPoints ... Input channels that other components or routines can write to
	EntryPoints() []chan models.TransitData
}

type metaData struct {
	id    models.ComponentID
	cType models.ComponentType

	*router
}

func (meta *metaData) ID() models.ComponentID {
	return meta.id
}

func (meta *metaData) Type() models.ComponentType {
	return meta.cType
}

type ComponentOption = func(*metaData)

func WithRouter(router *router) ComponentOption {
	return func(meta *metaData) {
		meta.router = router
	}
}

func WithID(id models.ComponentID) ComponentOption {
	return func(meta *metaData) {
		meta.id = id
	}
}
