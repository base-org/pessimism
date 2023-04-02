package pipeline

import (
	"github.com/base-org/pessimism/internal/conduit/models"
)

// Component ... Generalized interface that all pipeline components must adhere to
type Component interface {
	// Routing functionality for downstream communication
	AddDirective(id int, outChan chan models.TransitData) error
	RemoveDirective(id int) error

	// EventLoop ... Component driver function; spun up as separate go routine
	EventLoop() error
	Type() models.ComponentType
}
