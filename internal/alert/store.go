package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// AlertStore ... Interface for alert store
// NOTE - This is a simple in-memory store, using this interface
// we can easily swap it out for a persistent store
type AlertStore interface {
	AddAlertDestination(core.InvSessionUUID, core.AlertDestination) error
	GetAlertDestination(sUUID core.InvSessionUUID) (core.AlertDestination, error)
}

// AlertStore ... Alert store implementation
type alertStore struct {
	invariantAlertStore map[core.InvSessionUUID]core.AlertDestination
}

// NewAlertStore ... Initializer
func NewAlertStore() AlertStore {
	return &alertStore{
		invariantAlertStore: make(map[core.InvSessionUUID]core.AlertDestination),
	}
}

// AddAlertDestination ... Adds an alert destination for the given invariant session UUID
// NOTE - There can only be one alert destination per invariant session UUID
func (am *alertStore) AddAlertDestination(sUUID core.InvSessionUUID,
	alertDestination core.AlertDestination) error {
	if _, exists := am.invariantAlertStore[sUUID]; exists {
		return fmt.Errorf("alert destination already exists for invariant session %s", sUUID.String())
	}

	am.invariantAlertStore[sUUID] = alertDestination
	return nil
}

// GetAlertDestination ... Returns the alert destination for the given invariant session UUID
func (am *alertStore) GetAlertDestination(sUUID core.InvSessionUUID) (core.AlertDestination, error) {
	alertDestination, exists := am.invariantAlertStore[sUUID]
	if !exists {
		return 0, fmt.Errorf("alert destination does not exist for invariant session %s", sUUID.String())
	}

	return alertDestination, nil
}
