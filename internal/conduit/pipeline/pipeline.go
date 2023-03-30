package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/config"
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

// OracleConstructor ... Type declaration that a registry oracle component constructor must adhere to
type OracleConstructor = func(ctx context.Context, ot OracleType, cfg *config.OracleConfig) (Component, error)

// PipeConstructorFunc ... Type declaration that a registry pipe component constructor must adhere to
type PipeConstructorFunc = func(ctx context.Context, inputChan chan models.TransitData) (Component, error)
