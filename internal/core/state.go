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

func (sk *StateKey) SetPathID(PathID PathID) error {
	if sk.PathID != nil {
		return fmt.Errorf("state key already has a pipeline UUID %s", sk.PathID.String())
	}

	sk.PathID = &PathID
	return nil
}

func (sk StateKey) String() string {
	PathID := ""

	if sk.PathID != nil {
		PathID = sk.PathID.String()
	}

	return fmt.Sprintf("%s-%s-%s", PathID, sk.Prefix, sk.ID)
}
