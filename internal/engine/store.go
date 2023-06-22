package engine

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

// SessionStore ...
type SessionStore interface {
	AddInvSession(sUUID core.SUUID, pID core.PUUID, inv invariant.Invariant) error
	GetInvSessionByUUID(sUUID core.SUUID) (invariant.Invariant, error)
	GetInvariantsByUUIDs(sUUIDs ...core.SUUID) ([]invariant.Invariant, error)
	GetInvSessionsForPipeline(pUUID core.PUUID) ([]core.SUUID, error)
}

// sessionStore ...
type sessionStore struct {
	sessionPipelineMap map[core.PUUID][]core.SUUID
	invSessionMap      map[core.SUUID]invariant.Invariant // no duplicates
}

// NewSessionStore ... Initializer
func NewSessionStore() SessionStore {
	return &sessionStore{
		invSessionMap:      make(map[core.SUUID]invariant.Invariant),
		sessionPipelineMap: make(map[core.PUUID][]core.SUUID),
	}
}

// GetInvariantsByUUIDs ... Fetches in-order all invariants associated with a set of session UUIDs
func (ss *sessionStore) GetInvariantsByUUIDs(sUUIDs ...core.SUUID) ([]invariant.Invariant, error) {
	invariants := make([]invariant.Invariant, len(sUUIDs))

	for i, uuid := range sUUIDs {
		session, err := ss.GetInvSessionByUUID(uuid)
		if err != nil {
			return nil, err
		}

		invariants[i] = session
	}

	return invariants, nil
}

// GetInvSessionByUUID .... Fetches invariant session by UUID
func (ss *sessionStore) GetInvSessionByUUID(sUUID core.SUUID) (invariant.Invariant, error) {
	if entry, found := ss.invSessionMap[sUUID]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("invariant UUID doesn't exists in store inv mapping")
}

// GetInvSessionsForPipeline ... Returns all invariant session ids associated with pipeline
func (ss *sessionStore) GetInvSessionsForPipeline(pUUID core.PUUID) ([]core.SUUID, error) {
	if sessionIDs, found := ss.sessionPipelineMap[pUUID]; found {
		return sessionIDs, nil
	}
	return nil, fmt.Errorf("pipeline UUID doesn't exists in store inv mapping")
}

// AddInvSession ... Adds an invariant session to the store
func (ss *sessionStore) AddInvSession(sUUID core.SUUID,
	pUUID core.PUUID, inv invariant.Invariant) error {
	if _, found := ss.invSessionMap[sUUID]; found {
		return fmt.Errorf("invariant UUID already exists in store pid mapping")
	}

	if _, found := ss.sessionPipelineMap[pUUID]; !found { //
		ss.sessionPipelineMap[pUUID] = make([]core.SUUID, 0)
	}
	ss.invSessionMap[sUUID] = inv
	ss.sessionPipelineMap[pUUID] = append(ss.sessionPipelineMap[pUUID], sUUID)
	return nil
}

// RemoveInvSession ... Removes an existing invariant session from the store
func (ss *sessionStore) RemoveInvSession(_ core.SUUID,
	_ core.PUUID, _ invariant.Invariant) error {
	return nil
}
