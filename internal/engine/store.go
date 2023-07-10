package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

// SessionStore ...
type SessionStore interface {
	AddInvSession(sUUID core.SUUID, pID core.PUUID, inv invariant.Invariant) error
	GetInstanceByUUID(sUUID core.SUUID) (invariant.Invariant, error)
	GetInstancesByUUIDs(sUUIDs []core.SUUID) ([]invariant.Invariant, error)
	GetSUUIDsByPUUID(pUUID core.PUUID) ([]core.SUUID, error)
}

// sessionStore ...
type sessionStore struct {
	idMap       map[core.PUUID][]core.SUUID
	instanceMap map[core.SUUID]invariant.Invariant // no duplicates
}

// NewSessionStore ... Initializer
func NewSessionStore() SessionStore {
	return &sessionStore{
		instanceMap: make(map[core.SUUID]invariant.Invariant),
		idMap:       make(map[core.PUUID][]core.SUUID),
	}
}

// GetInstancesByUUIDs ... Fetches in-order all invariants associated with a set of session UUIDs
func (ss *sessionStore) GetInstancesByUUIDs(sUUIDs []core.SUUID) ([]invariant.Invariant, error) {
	invariants := make([]invariant.Invariant, len(sUUIDs))

	for i, uuid := range sUUIDs {
		session, err := ss.GetInstanceByUUID(uuid)
		if err != nil {
			return nil, err
		}

		invariants[i] = session
	}

	return invariants, nil
}

// GetInstanceByUUID .... Fetches invariant session by SUUID
func (ss *sessionStore) GetInstanceByUUID(sUUID core.SUUID) (invariant.Invariant, error) {
	if entry, found := ss.instanceMap[sUUID]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("invariant UUID doesn't exists in store inv mapping")
}

// GetSUUIDsByPUUID ... Returns all invariant session ids associated with pipeline
func (ss *sessionStore) GetSUUIDsByPUUID(pUUID core.PUUID) ([]core.SUUID, error) {
	if sessionIDs, found := ss.idMap[pUUID]; found {
		return sessionIDs, nil
	}
	return nil, fmt.Errorf("pipeline UUID doesn't exists in store inv mapping")
}

// AddInvSession ... Adds an invariant session to the store
func (ss *sessionStore) AddInvSession(sUUID core.SUUID,
	pUUID core.PUUID, inv invariant.Invariant) error {
	if _, found := ss.instanceMap[sUUID]; found {
		return fmt.Errorf("invariant UUID already exists in store pid mapping")
	}

	if _, found := ss.idMap[pUUID]; !found {
		ss.idMap[pUUID] = make([]core.SUUID, 0)
	}

	ss.instanceMap[sUUID] = inv
	ss.idMap[pUUID] = append(ss.idMap[pUUID], sUUID)
	return nil
}

// RemoveInvSession ... Removes an existing invariant session from the store
func (ss *sessionStore) RemoveInvSession(_ core.SUUID,
	_ core.PUUID, _ invariant.Invariant) error {
	return nil
}
