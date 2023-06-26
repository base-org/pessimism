package state

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// Store ... Interface for a state store
// TODO() - Add optional redis store implementation
type Store interface {
	GetSlice(context.Context, *core.StateKey) ([]string, error)
	GetNestedSubset(ctx context.Context, key *core.StateKey) (map[string][]string, error)

	SetSlice(context.Context, *core.StateKey, string) (string, error)
	Remove(context.Context, *core.StateKey) error
}

// FromContext ... Fetches a state store from context
func FromContext(ctx context.Context) (Store, error) {
	if store, ok := ctx.Value(core.State).(Store); ok {
		return store, nil
	}

	return nil, fmt.Errorf("could not load state object from context")
}
