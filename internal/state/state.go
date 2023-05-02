package state

import (
	"context"
	"fmt"
)

type StateKey int

const (
	Default StateKey = iota
)

type State interface {
	Get(ctx context.Context, key string) ([]string, error)
	Set(ctx context.Context, key string, value string) (string, error)
	Remove(ctx context.Context, key string) error
}

func FromContext(ctx context.Context) (State, error) {
	if store, ok := ctx.Value(Default).(State); ok {
		return store, nil
	}

	return nil, fmt.Errorf("could not load state object from context")
}
