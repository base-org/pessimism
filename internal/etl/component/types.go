package component

import (
	"context"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
)

type ActivityState int

const (
	Inactive ActivityState = iota
	Live
	Terminated
)

func (as ActivityState) String() string {
	switch as {
	case Inactive:
		return "inactive"

	case Live:
		return "live"

	case Terminated:
		return "terminated"
	}

	return "unknown"
}

// Router specific errors
const (
	dirAlreadyExistsErr = "%s directive key already exists within component router mapping"
	dirNotFoundErr      = "no directive key %s exists within component router mapping"

	transitErr = "[%s][%s] Received transit error: %s"
)

// Ingress specific errors
const (
	entryAlreadyExistsErr = "entrypoint already exists for %s"
	entryNotFoundErr      = "entrypoint not found for %s"
)

// Generalized component constructor types
type (
	// OracleConstructorFunc ... Type declaration that a registry oracle component constructor must adhere to
	OracleConstructorFunc = func(context.Context, core.PipelineType, *config.OracleConfig, ...Option) (Component, error)

	// PipeConstructorFunc ... Type declaration that a registry pipe component constructor must adhere to
	PipeConstructorFunc = func(context.Context, ...Option) (Component, error)
)
