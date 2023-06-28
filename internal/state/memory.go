package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

const (
	valAlreadySetError = "value already exists in state store"
	notFoundError      = "could not find state store value for key %s"
)

// IsValAlreadySetError ... Checks if the error is a ValAlreadySetError
func isValAlreadySetError(err error) bool {
	return err.Error() == valAlreadySetError
}

/*
	NOTE - This is a temporary implementation of the state store.
*/

// stateStore ... In memory state store
type stateStore struct {
	// NOTE - This is a temporary implementation of the state store.
	// Using a map of string to string slices to represent the state
	// store is not a scalable solution and will be rather expensive
	// in both memory and time complexity. This will be replaced with
	// a more optimal in-memory solution in the future.
	sliceStore map[string][]string

	sync.RWMutex
}

// NewMemState ... Initializer
func NewMemState() Store {
	return &stateStore{
		sliceStore: make(map[string][]string, 0),
		RWMutex:    sync.RWMutex{},
	}
}

// Get ... Fetches a string value slice from the store
func (ss *stateStore) GetSlice(_ context.Context, key *core.StateKey) ([]string, error) {
	ss.RLock()
	defer ss.RUnlock()

	val, exists := ss.sliceStore[key.String()]
	if !exists {
		return []string{}, fmt.Errorf(notFoundError, key)
	}

	return val, nil
}

// SetSlice ... Appends a value to the store slice
func (ss *stateStore) SetSlice(_ context.Context, key *core.StateKey, value string) (string, error) {
	ss.Lock()
	defer ss.Unlock()

	entries := ss.sliceStore[key.String()]
	for _, entry := range entries {
		if entry == value {
			return "", fmt.Errorf(valAlreadySetError)
		}
	}
	ss.sliceStore[key.String()] = append(entries, value)
	return value, nil
}

// Remove ... Removes a key entry from the store
func (ss *stateStore) Remove(_ context.Context, key *core.StateKey) error {
	ss.Lock()
	defer ss.Unlock()

	delete(ss.sliceStore, key.String())
	return nil
}

// GetNestedSubset ... Fetches a subset of a nested slice provided a nested
// key/value pair (ie. filters the state object into a subset object that
// contains only the values that match the nested key/value pair)
func (ss *stateStore) GetNestedSubset(_ context.Context,
	key *core.StateKey) (map[string][]string, error) {
	ss.RLock()
	defer ss.RUnlock()

	values, exists := ss.sliceStore[key.String()]
	if !exists {
		return map[string][]string{}, fmt.Errorf(notFoundError, key)
	}

	var nestedMap = make(map[string][]string, 0)
	for _, val := range values {
		if _, exists := ss.sliceStore[val]; !exists {
			return map[string][]string{}, fmt.Errorf(notFoundError, val)
		}

		nestedValues := ss.sliceStore[val]
		nestedMap[val] = nestedValues
	}

	return nestedMap, nil
}
