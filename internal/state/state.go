package state

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type CtxKey uint8

const (
	Default CtxKey = iota
)

// Store ... Interface for a state store
// TODO() - Add optional redis store implementation
type Store interface {
	GetSlice(context.Context, core.StateKey) ([]string, error)
	GetNestedSubset(ctx context.Context, key core.StateKey) (map[string][]string, error)

	SetSlice(context.Context, core.StateKey, string) (string, error)
	Remove(context.Context, core.StateKey) error
}

// FromContext ... Fetches a state store from context
func FromContext(ctx context.Context) (Store, error) {
	if store, ok := ctx.Value(Default).(Store); ok {
		return store, nil
	}

	return nil, fmt.Errorf("could not load state object from context")
}

// MakeKey ... Creates a state key
func MakeKey(prefix core.RegisterType, key string, nesting bool) core.StateKey {
	return core.StateKey{
		Nested: nesting,
		Key:    key,
		Prefix: uint8(prefix),
	}
}
