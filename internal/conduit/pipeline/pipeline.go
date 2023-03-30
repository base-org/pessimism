package pipeline

import (
	"github.com/base-org/pessimism/internal/conduit/models"
)

type PipelineComponent interface {
	AddDirective(id int, outChan chan models.TransitData) error
	RemoveDirective(id int) error

	// EventLoop - Component driver function
	EventLoop() error
	Type() models.ComponentType
}
