//go:generate mockgen -package mocks --destination ../../mocks/alert_client.go . AlertClient

package client

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
)

// AlertClient ... An interface for alert clients to implement
type AlertClient interface {
	PostEvent(ctx context.Context, data *AlertEventTrigger) (*AlertAPIResponse, error)
	GetName() string
}

// AlertEventTrigger ... A standardized event trigger for alert clients
type AlertEventTrigger struct {
	Message  string
	Severity core.Severity
	DedupKey core.PUUID
}

// AlertAPIResponse ... A standardized response for alert clients
type AlertAPIResponse struct {
	Status  core.AlertStatus
	Message string
}

// ToPagerdutyEvent ... Converts an AlertEventTrigger to a PagerDutyEventTrigger
func (a *AlertEventTrigger) ToPagerdutyEvent() *PagerDutyEventTrigger {
	return &PagerDutyEventTrigger{
		DedupKey: a.DedupKey.String(),
		Severity: a.Severity.ToPagerDutySev(),
		Message:  a.Message,
	}
}
