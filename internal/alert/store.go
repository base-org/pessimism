package alert

import (
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

// Store ... Interface for alert policy store
// NOTE - This is a simple in-memory store, using this interface
// we can easily swap it out for a persistent store
type Store interface {
	AddAlertPolicy(core.SUUID, *core.AlertPolicy) error
	GetAlertPolicy(sUUID core.SUUID) (*core.AlertPolicy, error)
}

// store ... Alert store implementation
// Used to store critical alerting metadata (ie. alert destination, message, etc.)
type store struct {
	defMap           map[core.SUUID]*core.AlertPolicy
	pagerDutyClients map[core.Severity][]client.PagerDutyClient
	slackClients     map[core.Severity][]client.SlackClient
	cfg              *Config
}

// NewStore ... Initializer
func NewStore(cfg *Config) Store {
	return &store{
		cfg:              cfg,
		defMap:           make(map[core.SUUID]*core.AlertPolicy),
		pagerDutyClients: make(map[core.Severity][]client.PagerDutyClient),
		slackClients:     make(map[core.Severity][]client.SlackClient),
	}
}

// AddAlertPolicy ... Adds an alert policy for the given heuristic session UUID
// NOTE - There can only be one alert destination per heuristic session UUID
func (am *store) AddAlertPolicy(sUUID core.SUUID, policy *core.AlertPolicy) error {
	if _, exists := am.defMap[sUUID]; exists {
		return fmt.Errorf("alert destination already exists for heuristic session %s", sUUID.String())
	}

	am.defMap[sUUID] = policy
	return nil
}

// GetAlertPolicy ... Returns the alert destination for the given heuristic session UUID
func (am *store) GetAlertPolicy(sUUID core.SUUID) (*core.AlertPolicy, error) {
	dest, exists := am.defMap[sUUID]
	if !exists {
		return nil, fmt.Errorf("alert destination does not exist for heuristic session %s", sUUID.String())
	}

	return dest, nil
}
