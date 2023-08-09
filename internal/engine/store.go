package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
)

// SessionStore ...
type SessionStore interface {
	AddSession(sUUID core.SUUID, pID core.PUUID, h heuristic.Heuristic) error
	GetInstanceByUUID(sUUID core.SUUID) (heuristic.Heuristic, error)
	GetInstancesByUUIDs(sUUIDs []core.SUUID) ([]heuristic.Heuristic, error)
	GetSUUIDsByPUUID(pUUID core.PUUID) ([]core.SUUID, error)
}

// sessionStore ...
type sessionStore struct {
	idMap       map[core.PUUID][]core.SUUID
	instanceMap map[core.SUUID]heuristic.Heuristic // no duplicates
}

// NewSessionStore ... Initializer
func NewSessionStore() SessionStore {
	return &sessionStore{
		instanceMap: make(map[core.SUUID]heuristic.Heuristic),
		idMap:       make(map[core.PUUID][]core.SUUID),
	}
}

// GetInstancesByUUIDs ... Fetches in-order all heuristics associated with a set of session UUIDs
func (ss *sessionStore) GetInstancesByUUIDs(sUUIDs []core.SUUID) ([]heuristic.Heuristic, error) {
	heuristics := make([]heuristic.Heuristic, len(sUUIDs))

	for i, uuid := range sUUIDs {
		session, err := ss.GetInstanceByUUID(uuid)
		if err != nil {
			return nil, err
		}

		heuristics[i] = session
	}

	return heuristics, nil
}

// GetInstanceByUUID .... Fetches heuristic session by SUUID
func (ss *sessionStore) GetInstanceByUUID(sUUID core.SUUID) (heuristic.Heuristic, error) {
	if entry, found := ss.instanceMap[sUUID]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("heuristic UUID doesn't exists in store heuristic mapping")
}

// GetSUUIDsByPUUID ... Returns all heuristic session ids associated with pipeline
func (ss *sessionStore) GetSUUIDsByPUUID(pUUID core.PUUID) ([]core.SUUID, error) {
	if sessionIDs, found := ss.idMap[pUUID]; found {
		return sessionIDs, nil
	}
	return nil, fmt.Errorf("pipeline UUID doesn't exists in store heuristic mapping")
}

// AddSession ... Adds a heuristic session to the store
func (ss *sessionStore) AddSession(sUUID core.SUUID,
	pUUID core.PUUID, h heuristic.Heuristic) error {
	if _, found := ss.instanceMap[sUUID]; found {
		return fmt.Errorf("heuristic UUID already exists in store pid mapping")
	}

	if _, found := ss.idMap[pUUID]; !found {
		ss.idMap[pUUID] = make([]core.SUUID, 0)
	}

	ss.instanceMap[sUUID] = h
	ss.idMap[pUUID] = append(ss.idMap[pUUID], sUUID)
	return nil
}

// RemoveInvSession ... Removes an existing heuristic session from the store
func (ss *sessionStore) RemoveInvSession(_ core.SUUID,
	_ core.PUUID, _ heuristic.Heuristic) error {
	return nil
}
