package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// Store ... Interface for alert policy store
// NOTE - This is a simple in-memory store, using this interface
// we can easily swap it out for a persistent store
type Store interface {
	AddAlertPolicy(core.UUID, *core.AlertPolicy) error
	GetAlertPolicy(sUUID core.UUID) (*core.AlertPolicy, error)
}

// store ... Alert store implementation
// Used to store critical alerting metadata (ie. alert destination, message, etc.)
type store struct {
	defMap map[core.UUID]*core.AlertPolicy
}

// NewStore ... Initializer
func NewStore() Store {
	return &store{
		defMap: make(map[core.UUID]*core.AlertPolicy),
	}
}

// AddAlertPolicy ... Adds an alert policy for the given heuristic session UUID
// NOTE - There can only be one alert destination per heuristic session UUID
func (am *store) AddAlertPolicy(sUUID core.UUID, policy *core.AlertPolicy) error {
	if _, exists := am.defMap[sUUID]; exists {
		return fmt.Errorf("alert destination already exists for heuristic session %s", sUUID.String())
	}

	am.defMap[sUUID] = policy
	return nil
}

// GetAlertPolicy ... Returns the alert destination for the given heuristic session UUID
func (am *store) GetAlertPolicy(sUUID core.UUID) (*core.AlertPolicy, error) {
	dest, exists := am.defMap[sUUID]
	if !exists {
		return nil, fmt.Errorf("alert destination does not exist for heuristic session %s", sUUID.String())
	}

	return dest, nil
}
