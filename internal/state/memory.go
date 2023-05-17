package state

import (
	"context"
	"fmt"
	"sync"
)

type stateStore struct {
	store map[string][]string

	sync.RWMutex
}

func NewMemState() StateDB {

	return &stateStore{
		store: make(map[string][]string, 0),

		RWMutex: sync.RWMutex{},
	}
}

func (ss *stateStore) Get(ctx context.Context, key string) ([]string, error) {
	ss.RLock()
	defer ss.RUnlock()

	val, exists := ss.store[key]
	if !exists {
		return []string{}, fmt.Errorf("could not find value")
	}

	return val, nil
}

func (ss *stateStore) Set(ctx context.Context, key string, value string) (string, error) {
	ss.Lock()
	defer ss.Unlock()

	ss.store[key] = append(ss.store[key], value)

	return value, nil
}

func (ss *stateStore) Remove(ctx context.Context, key string) error {
	ss.Lock()
	defer ss.Unlock()

	delete(ss.store, key)
	return nil
}
