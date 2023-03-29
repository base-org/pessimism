package pipeline

import (
	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/rs/xid"
)

type PipelineComponent interface {
	AddDirective(id xid.ID, outChan chan models.TransitData) error
	RemoveDirective(id xid.ID) error

	// EventLoop - Component driver function
	EventLoop() error
	Type() models.ComponentType
}
