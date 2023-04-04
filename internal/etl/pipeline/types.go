package pipeline

import (
	"context"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/models"
)

// Router specific errors
const (
	dirAlreadyExistsErr = "%d directive key already exists within component router mapping"
	dirNotFoundErr      = "no directive key %d exists within component router mapping"
)

// Generalized component constructor types
type (
	// OracleConstructorFunc ... Type declaration that a registry oracle component constructor must adhere to
	OracleConstructorFunc = func(context.Context, models.PipelineType, *config.OracleConfig) (Component, error)

	// PipeConstructorFunc ... Type declaration that a registry pipe component constructor must adhere to
	PipeConstructorFunc = func(context.Context) (Component, error)
)
