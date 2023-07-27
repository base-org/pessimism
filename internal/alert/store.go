package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// TODO(#81): No Support for Multiple Alerting Destinations for an Heuristic Session

// Store ... Interface for alert store
// NOTE - This is a simple in-memory store, using this interface
// we can easily swap it out for a persistent store
type Store interface {
	AddAlertDestination(core.SUUID, core.AlertDestination) error
	GetAlertDestination(sUUID core.SUUID) (core.AlertDestination, error)
}

// store ... Alert store implementation
type store struct {
	alertMap map[core.SUUID]core.AlertDestination
}

// NewStore ... Initializer
func NewStore() Store {
	return &store{
		alertMap: make(map[core.SUUID]core.AlertDestination),
	}
}

// AddAlertDestination ... Adds an alert destination for the given heuristic session UUID
// NOTE - There can only be one alert destination per heuristic session UUID
func (am *store) AddAlertDestination(sUUID core.SUUID,
	alertDestination core.AlertDestination) error {
	if _, exists := am.alertMap[sUUID]; exists {
		return fmt.Errorf("alert destination already exists for heuristic session %s", sUUID.String())
	}

	am.alertMap[sUUID] = alertDestination
	return nil
}

// GetAlertDestination ... Returns the alert destination for the given heuristic session UUID
func (am *store) GetAlertDestination(sUUID core.SUUID) (core.AlertDestination, error) {
	dest, exists := am.alertMap[sUUID]
	if !exists {
		return 0, fmt.Errorf("alert destination does not exist for heuristic session %s", sUUID.String())
	}

	return dest, nil
}
