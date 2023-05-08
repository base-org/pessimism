package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

type InvariantStore struct {
	pidMap map[core.RegisterPID][]invariant.Invariant // duplicates allowed
	invMap map[core.InvariantUUID]invariant.Invariant // no duplicates
}

func NewInvariantStore() *InvariantStore {
	return &InvariantStore{
		pidMap: make(map[core.RegisterPID][]invariant.Invariant),
		invMap: make(map[core.InvariantUUID]invariant.Invariant),
	}
}

func (is *InvariantStore) GetInvariantsByRegisterPID(id core.RegisterPID) ([]invariant.Invariant, error) {
	return is.pidMap[id], nil
}

func (is *InvariantStore) GetInvariantByUUID(id core.InvariantUUID) (invariant.Invariant, error) {
	if entry, found := is.invMap[id]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("invariant UUID doesn't exists in store inv mapping")
}

func (is *InvariantStore) AddInvariant(id core.InvariantUUID, rid core.RegisterPID, inv invariant.Invariant) error {
	if _, found := is.invMap[id]; found {
		return fmt.Errorf("invariant UUID already exists in store pid mapping")
	}

	if _, found := is.pidMap[rid]; !found {
		is.pidMap[rid] = make([]invariant.Invariant, 0)
	}

	is.pidMap[rid] = append(is.pidMap[rid], inv)
	is.invMap[id] = inv
	return nil
}
