package pipeline

import "github.com/base-org/pessimism/internal/conduit/models"

type PipelineComponent interface {
	// EventLoop - Component driver function
	EventLoop() error
	Type() models.ComponentType
}
