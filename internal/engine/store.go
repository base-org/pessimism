package engine

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/stretchr/testify/assert"
)

// SessionStore ...
type SessionStore interface {
	AddInvSession(sUUID core.InvSessionUUID, pID core.PipelineUUID, inv invariant.Invariant) error
	GetInvSessionByUUID(sUUID core.InvSessionUUID) (invariant.Invariant, error)
	GetInvariantsByUUIDs(sUUIDs ...core.InvSessionUUID) ([]invariant.Invariant, error)
	GetInvSessionsForPipeline(pUUID core.PipelineUUID) ([]core.InvSessionUUID, error)
}

// sessionStore ...
type sessionStore struct {
	sessionPipelineMap map[core.PipelineUUID][]core.InvSessionUUID
	invSessionMap      map[core.InvSessionUUID]invariant.Invariant // no duplicates
}

// NewSessionStore ... Initializer
func NewSessionStore() SessionStore {
	return &sessionStore{
		invSessionMap:      make(map[core.InvSessionUUID]invariant.Invariant),
		sessionPipelineMap: make(map[core.PipelineUUID][]core.InvSessionUUID),
	}
}

// GetInvariantsByUUIDs ... Fetches in-order all invariants associated with a set of session UUIDs
func (ss *sessionStore) GetInvariantsByUUIDs(sUUIDs ...core.InvSessionUUID) ([]invariant.Invariant, error) {
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
func (ss *sessionStore) GetInvSessionByUUID(sUUID core.InvSessionUUID) (invariant.Invariant, error) {
	if entry, found := ss.invSessionMap[sUUID]; found {
		return entry, nil
	}
	return nil, fmt.Errorf("invariant UUID doesn't exists in store inv mapping")
}

// GetInvSessionsForPipeline ... Returns all invariant session ids associated with pipeline
func (ss *sessionStore) GetInvSessionsForPipeline(pUUID core.PipelineUUID) ([]core.InvSessionUUID, error) {
	if sessionIDs, found := ss.sessionPipelineMap[pUUID]; found {
		return sessionIDs, nil
	}
	return nil, fmt.Errorf("pipeline UUID doesn't exists in store inv mapping")
}

// AddInvSession ... Adds an invariant session to the store
func (ss *sessionStore) AddInvSession(sUUID core.InvSessionUUID,
	pUUID core.PipelineUUID, inv invariant.Invariant) error {
	if _, found := ss.invSessionMap[sUUID]; found {
		return fmt.Errorf("invariant UUID already exists in store pid mapping")
	}

	if _, found := ss.sessionPipelineMap[pUUID]; !found { //
		ss.sessionPipelineMap[pUUID] = make([]core.InvSessionUUID, 0)
	}
	ss.invSessionMap[sUUID] = inv
	ss.sessionPipelineMap[pUUID] = append(ss.sessionPipelineMap[pUUID], sUUID)
	return nil
}

// RemoveInvSession ... Removes an existing invariant session from the store
func (ss *sessionStore) RemoveInvSession(_ core.InvSessionUUID,
	_ core.PipelineUUID, _ invariant.Invariant) error {
	return nil
}

func TestSessionStore(t *testing.T) {
	// Setup
	ss := NewSessionStore()

	// Test GetInvSessionByUUID
	_, err := ss.GetInvSessionByUUID(core.NilInvariantUUID())
	assert.Error(t, err, "should error")
}
