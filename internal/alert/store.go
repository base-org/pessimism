package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type AlertStore interface {
	AddAlertDestination(core.InvSessionUUID, core.AlertDestination) error
	GetAlertDestination(sUUID core.InvSessionUUID) (core.AlertDestination, error)
}

type alertStore struct {
	invariantAlertStore map[core.InvSessionUUID]core.AlertDestination
}

func NewAlertStore() AlertStore {
	return &alertStore{
		invariantAlertStore: make(map[core.InvSessionUUID]core.AlertDestination),
	}
}

func (am *alertStore) AddAlertDestination(sUUID core.InvSessionUUID,
	alertDestination core.AlertDestination) error {
	if _, exists := am.invariantAlertStore[sUUID]; exists {
		return fmt.Errorf("alert destination already exists for invariant session %s", sUUID.String())
	}

	am.invariantAlertStore[sUUID] = alertDestination
	return nil
}

func (am *alertStore) GetAlertDestination(sUUID core.InvSessionUUID) (core.AlertDestination, error) {
	alertDestination, exists := am.invariantAlertStore[sUUID]
	if !exists {
		return 0, fmt.Errorf("alert destination does not exist for invariant session %s", sUUID.String())
	}

	return alertDestination, nil
}
