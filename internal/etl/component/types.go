package component

import (
	"context"

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

// StateChange ... Represents a component state change event
// that is processed by component management logic to determine
// proper pipeline states and duplicate pipeline merging opportunities
type StateChange struct {
	ID core.CUUID

	From ActivityState // S
	To   ActivityState // S'
}

// EgressHandler specific errors
const (
	engineEgressExistsErr  = "engine egress already exists"
	egressAlreadyExistsErr = "%s egress key already exists within component router mapping"
	egressNotFoundErr      = "no egress key %s exists within component router mapping"
	egressNotExistErr      = "received transit request with 0 out channels to write to"

	transitErr = "received transit error: %s"
)

// IngressHandler specific errors
const (
	ingressAlreadyExistsErr = "ingress already exists for %s"
	ingressNotFoundErr      = "ingress not found for %s"
)

type (
	// OracleConstructorFunc ... Type declaration that a registry oracle component constructor must adhere to
	OracleConstructorFunc = func(context.Context, *core.ClientConfig, ...Option) (Component, error)

	// PipeConstructorFunc ... Type declaration that a registry pipe component constructor must adhere to
	PipeConstructorFunc = func(context.Context, *core.ClientConfig, ...Option) (Component, error)
)

// OracleType ...
type OracleType = string

const (
	// BackTestOracle ... Represents an oracle used for backtesting some invariant
	BacktestOracle OracleType = "backtest"
	// LiveOracle ... Represents an oracle used for powering some live invariant
	LiveOracle OracleType = "live"
)
