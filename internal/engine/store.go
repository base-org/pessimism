package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

type InvariantStore struct {
	invPipeLineMap map[core.PipelineUUID][]core.InvSessionUUID
	invMap         map[core.InvSessionUUID]invariant.Invariant // no duplicates
}

func NewInvariantStore() *InvariantStore {
	return &InvariantStore{
		invMap:         make(map[core.InvSessionUUID]invariant.Invariant),
		invPipeLineMap: make(map[core.PipelineUUID][]core.InvSessionUUID),
	}
}

func (is *InvariantStore) GetInvSessionByUUID(id core.InvSessionUUID) (invariant.Invariant, error) {
	if entry, found := is.invMap[id]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("invariant UUID doesn't exists in store inv mapping")
}

func (is *InvariantStore) AddInvSession(id core.InvSessionUUID, pID core.PipelineUUID, inv invariant.Invariant) error {
	if _, found := is.invMap[id]; found {
		return fmt.Errorf("invariant UUID already exists in store pid mapping")
	}

	if _, found := is.invPipeLineMap[pID]; !found {
		is.invPipeLineMap[pID] = make([]core.InvSessionUUID, 0)
	}
	is.invMap[id] = inv
	is.invPipeLineMap[pID] = append(is.invPipeLineMap[pID], id)
	return nil
}
