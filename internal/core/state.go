package core

import (
	"fmt"
)

type StateKey struct {
	Nesting bool
	Prefix  TopicType
	ID      string

	PathID *PathID
}

func (sk *StateKey) Clone() *StateKey {
	return &StateKey{
		Nesting: sk.Nesting,
		Prefix:  sk.Prefix,
		ID:      sk.ID,
		PathID:  sk.PathID,
	}
}

func MakeStateKey(pre TopicType, id string, nest bool) *StateKey {
	return &StateKey{
		Nesting: nest,
		Prefix:  pre,
		ID:      id,
	}
}

func (sk *StateKey) IsNested() bool {
	return sk.Nesting
}

func (sk *StateKey) SetPathID(id PathID) error {
	if sk.PathID != nil {
		return fmt.Errorf("state key already has a path UUID %s", sk.PathID.String())
	}

	sk.PathID = &id
	return nil
}

func (sk StateKey) String() string {
	id := ""

	if sk.PathID != nil {
		id = sk.PathID.String()
	}

	return fmt.Sprintf("%s-%s-%s", id, sk.Prefix, sk.ID)
}
