//go:generate mockgen -package mocks --destination ../mocks/pagerduty_client.go . PagerdutyClient

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/base-org/pessimism/internal/logging"

	"go.uber.org/zap"
)

// PagerDutyAction represents the type of actions that can be triggered by an event
type PagerDutyAction string

const (
	Trigger                PagerDutyAction = "trigger"
	PagerDutyAckAction     PagerDutyAction = "acknowledge"
	PagerDutyResolveAction PagerDutyAction = "resolve"
)

// PagerDutySeverity represents the severity of an event
type PagerDutySeverity string

const (
	Critical PagerDutySeverity = "critical"
	Error    PagerDutySeverity = "error"
	Warning  PagerDutySeverity = "warning"
	Info     PagerDutySeverity = "info"
)

// PagerDutyResponseStatus is the response status of a PagerDuty API call
type PagerDutyResponseStatus string

const (
	SuccessStatus PagerDutyResponseStatus = "success"
)

// PagerDutyConfig ... Represents the configuration vars for a PagerDuty client
type PagerDutyConfig struct {
	IntegrationKey  string
	ChangeEventsURL string
	AlertEventsURL  string
}

// PagerDutyClient ... Interface for PagerDuty client
type PagerDutyClient interface {
	PostEvent(ctx context.Context, event *PagerDutyEventTrigger) (*PagerDutyAPIResponse, error)
}

// pagerdutyClient ... PagerDuty client for making requests
type pagerdutyClient struct {
	integrationKey  string
	changeEventsURL string
	alertEventsURL  string
	client          *http.Client
}

// NewPagerDutyClient ... Initializer for PagerDuty client
func NewPagerDutyClient(cfg *PagerDutyConfig) PagerDutyClient {
	if cfg.IntegrationKey == "" {
		logging.NoContext().Warn("No PagerDuty integration key provided")
	}

	return &pagerdutyClient{
		integrationKey:  cfg.IntegrationKey,
		changeEventsURL: cfg.ChangeEventsURL,
		alertEventsURL:  cfg.AlertEventsURL,
		client:          &http.Client{},
	}
}

// PagerDutyEventTrigger ... Represents caller specified fields for a PagerDuty event
type PagerDutyEventTrigger struct {
	Message  string
	Action   PagerDutyAction
	Severity PagerDutySeverity
	DedupKey string
}

// PagerDutyRequest ... Used to construct a PagerDuty api request
type PagerDutyRequest struct {
	RoutingKey  string           `json:"routing_key"`
	EventAction PagerDutyAction  `json:"event_action"`
	DedupKey    string           `json:"dedup_key"`
	Payload     PagerDutyPayload `json:"payload"`
}

// PagerDutyPayload ... Represents the payload of a PagerDuty event
type PagerDutyPayload struct {
	Summary   string            `json:"summary"`
	Source    string            `json:"source"`
	Severity  PagerDutySeverity `json:"severity"`
	Timestamp time.Time         `json:"timestamp"`
}

// newPagerDutyPayload ... Initializes a new PagerDuty payload given the integration key and event
func newPagerDutyPayload(integrationKey string, event *PagerDutyEventTrigger) *PagerDutyRequest {
	return &PagerDutyRequest{
		RoutingKey:  integrationKey,
		EventAction: event.Action,
		DedupKey:    event.DedupKey,
		Payload: PagerDutyPayload{
			Summary:   event.Message,
			Source:    "Pessimism",
			Severity:  event.Severity,
			Timestamp: time.Now(),
		},
	}
}

// marshal ... Marshals the PagerDuty payload
func (req *PagerDutyRequest) marshal() ([]byte, error) {
	bytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// PagerDutyAPIResponse ... Represents the structure of a PagerDuty API response
type PagerDutyAPIResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	DedupKey string `json:"dedup_key"`
}

// PostEvent ... Posts a new event to PagerDuty
func (pdc pagerdutyClient) PostEvent(ctx context.Context, event *PagerDutyEventTrigger) (*PagerDutyAPIResponse, error) {
	// 1. Create and marshal payload into request object body

	if pdc.integrationKey == "" {
		return nil, fmt.Errorf("no Pagerduty integration key provided")
	}

	payload, err := newPagerDutyPayload(pdc.integrationKey, event).marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pdc.alertEventsURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 2. Make request to PagerDuty
	resp, err := pdc.client.Do(req)
	defer func() {
		if err = resp.Body.Close(); err != nil {
			logging.WithContext(ctx).Warn("Could not close pagerduty response body",
				zap.Error(err))
		}
	}()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp *PagerDutyAPIResponse
	if err := json.Unmarshal(bytes, &apiResp); err != nil {
		return nil, err
	}

	return apiResp, nil
}
