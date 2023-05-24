package state

import (
	"context"
	"fmt"
	"sync"
)

// stateStore ... In memory state store
type stateStore struct {
	store map[string][]string

	sync.RWMutex
}

// NewMemState ... Initializer
func NewMemState() Store {
	return &stateStore{
		store:   make(map[string][]string, 0),
		RWMutex: sync.RWMutex{},
	}
}

// Get ... Fetches a string value slice from the store
func (ss *stateStore) Get(_ context.Context, key string) ([]string, error) {
	ss.RLock()
	defer ss.RUnlock()

	val, exists := ss.store[key]
	if !exists {
		return []string{}, fmt.Errorf("could not find state store value for key %s", key)
	}

	return val, nil
}

// Set ... Appends a value to the store slice
func (ss *stateStore) Set(_ context.Context, key string, value string) (string, error) {
	ss.Lock()
	defer ss.Unlock()

	ss.store[key] = append(ss.store[key], value)

	return value, nil
}

// Remove ... Removes a key entry from the store
func (ss *stateStore) Remove(_ context.Context, key string) error {
	ss.Lock()
	defer ss.Unlock()

	delete(ss.store, key)
	return nil
}

// Merge ... Merges two keys together
func (ss *stateStore) Merge(_ context.Context, key1, key2 string) error {
	ss.Lock()
	defer ss.Unlock()

	ss.store[key2] = append(ss.store[key2], ss.store[key1]...)
	delete(ss.store, key1)

	return nil
}
