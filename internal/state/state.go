package state

import (
	"context"
	"fmt"
)

type Key int

const (
	Default Key = iota
)

// Store ... Interface for a state store
type Store interface {
	Get(ctx context.Context, key string) ([]string, error)
	Set(ctx context.Context, key string, value string) (string, error)
	Remove(ctx context.Context, key string) error
}

// FromContext ... Fetches a state store from context
func FromContext(ctx context.Context) (Store, error) {
	if store, ok := ctx.Value(Default).(Store); ok {
		return store, nil
	}

	return nil, fmt.Errorf("could not load state object from context")
}
