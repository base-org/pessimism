package state

import (
	"context"
	"fmt"
)

type StateKey int

const (
	Default StateKey = iota
)

type StateDB interface {
	Get(ctx context.Context, key string) ([]string, error)
	Set(ctx context.Context, key string, value string) (string, error)
	Remove(ctx context.Context, key string) error
}

func FromContext(ctx context.Context) (StateDB, error) {
	if store, ok := ctx.Value(Default).(StateDB); ok {
		return store, nil
	}

	return nil, fmt.Errorf("could not load state object from context")
}
