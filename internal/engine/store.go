package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
)

// SessionStore ...
type SessionStore interface {
	AddSession(sUUID core.UUID, pID core.PathID, h heuristic.Heuristic) error
	GetInstanceByUUID(sUUID core.UUID) (heuristic.Heuristic, error)
	GetInstancesByUUIDs(sUUIDs []core.UUID) ([]heuristic.Heuristic, error)
	GetUUIDsByPathID(PathID core.PathID) ([]core.UUID, error)
}

// sessionStore ...
type sessionStore struct {
	idMap       map[core.PathID][]core.UUID
	instanceMap map[core.UUID]heuristic.Heuristic // no duplicates
}

// NewSessionStore ... Initializer
func NewSessionStore() SessionStore {
	return &sessionStore{
		instanceMap: make(map[core.UUID]heuristic.Heuristic),
		idMap:       make(map[core.PathID][]core.UUID),
	}
}

func (ss *sessionStore) GetInstancesByUUIDs(sUUIDs []core.UUID) ([]heuristic.Heuristic, error) {
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

func (ss *sessionStore) GetInstanceByUUID(sUUID core.UUID) (heuristic.Heuristic, error) {
	if entry, found := ss.instanceMap[sUUID]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("heuristic UUID doesn't exists in store heuristic mapping")
}

func (ss *sessionStore) GetUUIDsByPathID(PathID core.PathID) ([]core.UUID, error) {
	if sessionIDs, found := ss.idMap[PathID]; found {
		return sessionIDs, nil
	}
	return nil, fmt.Errorf("path UUID doesn't exists in store heuristic mapping")
}

func (ss *sessionStore) AddSession(sUUID core.UUID,
	PathID core.PathID, h heuristic.Heuristic) error {
	if _, found := ss.instanceMap[sUUID]; found {
		return fmt.Errorf("heuristic UUID already exists in store pid mapping")
	}

	if _, found := ss.idMap[PathID]; !found {
		ss.idMap[PathID] = make([]core.UUID, 0)
	}

	ss.instanceMap[sUUID] = h
	ss.idMap[PathID] = append(ss.idMap[PathID], sUUID)
	return nil
}

func (ss *sessionStore) RemoveInvSession(_ core.UUID,
	_ core.PathID, _ heuristic.Heuristic) error {
	return nil
}
