package process

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

// Denotes a process state change
type StateChange struct {
	ID core.ProcessID

	From ActivityState // S
	To   ActivityState // S'
}

const (
	engineRelayExists = "engine egress already exists"
	subExistsErr      = "%s subscriber already exists"
	subNotFound       = "no subscriber with key %s exists"
	noSubErr          = "no subscribers to notify"

	relayErr = "received relay error: %s"
)

const (
	topicExistsErr   = "topic already exists for %s"
	topicNotFoundErr = "topic not found for %s"
)

type (
	Constructor = func(context.Context, *core.ClientConfig, ...Option) (Process, error)
)
