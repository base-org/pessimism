package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/base-org/pessimism/internal/core"
)

/*
	NOTE - This is a temporary implementation of the state store.
*/

// stateStore ... In memory state store
type stateStore struct {
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
func (ss *stateStore) GetSlice(_ context.Context, key core.StateKey) ([]string, error) {
	ss.RLock()
	defer ss.RUnlock()

	val, exists := ss.sliceStore[key.String()]
	if !exists {
		return []string{}, fmt.Errorf("could not find state store value for key %s", key)
	}

	return val, nil
}

// SetSlice ... Appends a value to the store slice
func (ss *stateStore) SetSlice(_ context.Context, key core.StateKey, value string) (string, error) {
	ss.Lock()
	defer ss.Unlock()

	ss.sliceStore[key.String()] = append(ss.sliceStore[key.String()], value)

	return value, nil
}

// Remove ... Removes a key entry from the store
func (ss *stateStore) Remove(_ context.Context, key core.StateKey) error {
	ss.Lock()
	defer ss.Unlock()

	delete(ss.sliceStore, key.String())
	return nil
}

// GetNestedSubset ... Fetches a subset of a nested slice provided a nested
// key/value pair (ie. filters the state object into a subset object that
// contains only the values that match the nested key/value pair)
func (ss *stateStore) GetNestedSubset(_ context.Context,
	key core.StateKey) (map[string][]string, error) {
	ss.RLock()
	defer ss.RUnlock()

	values, exists := ss.sliceStore[key.String()]
	if !exists {
		return map[string][]string{}, fmt.Errorf("could not find state store value for key %s", key)
	}

	var nestedMap = make(map[string][]string, 0)
	for _, val := range values {
		if _, exists := ss.sliceStore[val]; !exists {
			return map[string][]string{}, fmt.Errorf("could not find state store value for key %s", key)
		}

		nestedValues := ss.sliceStore[val]
		nestedMap[val] = nestedValues
	}

	return nestedMap, nil
}
