package core

import (
	"fmt"
)

// StateKey ... Represents a key in the state store
type StateKey struct {
	Nesting bool
	Prefix  RegisterType
	ID      string

	PUUID *PUUID
}

// Clone ... Returns a copy of the state key
func (sk *StateKey) Clone() *StateKey {
	return &StateKey{
		Nesting: sk.Nesting,
		Prefix:  sk.Prefix,
		ID:      sk.ID,
		PUUID:   sk.PUUID,
	}
}

// MakeStateKey ... Builds a minimal state key using
// a prefix and key
func MakeStateKey(pre RegisterType, id string, nest bool) *StateKey {
	return &StateKey{
		Nesting: nest,
		Prefix:  pre,
		ID:      id,
	}
}

// IsNested ... Indicates whether the state key is nested
// NOTE - This is used to determine if the state key maps
// to a value slice of state keys in the state store (ie. nested)
func (sk *StateKey) IsNested() bool {
	return sk.Nesting
}

// SetPUUID ... Adds a pipeline UUID to the state key prefix and returns a new state key
func (sk *StateKey) SetPUUID(pUUID PUUID) error {
	if sk.PUUID != nil {
		return fmt.Errorf("state key already has a pipeline UUID %s", sk.PUUID.String())
	}

	sk.PUUID = &pUUID
	return nil
}

const (
	AddressPrefix = iota + 1
	NestedPrefix
)

// String ... Returns a string representation of the state key
func (sk StateKey) String() string {
	pUUID := ""

	if sk.PUUID != nil {
		pUUID = sk.PUUID.String()
	}

	return fmt.Sprintf("%s-%s-%s", pUUID, sk.Prefix, sk.ID)
}
