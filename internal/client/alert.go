//go:generate mockgen -package mocks --destination ../mocks/alert_client.go . AlertClient

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
	Message string
	Alert   core.Alert
}

// AlertAPIResponse ... A standardized response for alert clients
type AlertAPIResponse struct {
	Status  core.AlertStatus
	Message string
}

// ToPagerdutyEvent ... Converts an AlertEventTrigger to a PagerDutyEventTrigger
func (a *AlertEventTrigger) ToPagerdutyEvent() *PagerDutyEventTrigger {
	return &PagerDutyEventTrigger{
		DedupKey: a.Alert.PathID.String(),
		Severity: a.Alert.Sev.ToPagerDutySev(),
		Message:  a.Message,
	}
}

func (a *AlertEventTrigger) ToSNSMessagePayload() *SNSMessagePayload {
	return &SNSMessagePayload{
		Network:       a.Alert.Net.String(),
		HeuristicType: a.Alert.HT.String(),
		Severity:      a.Alert.Sev.String(),
		PathID:        a.Alert.PathID.String(),
		HeuristicID:   a.Alert.HeuristicID.String(),
		Timestamp:     a.Alert.Timestamp,
		Content:       a.Message,
	}
}
