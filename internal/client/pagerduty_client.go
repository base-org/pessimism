//go:generate mockgen -package mocks --destination ../mocks/pagerduty_client.go . PagerdutyClient

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/base-org/pessimism/internal/logging"

	"go.uber.org/zap"
)

// PagerdutyAction represents the type of actions that can be triggered by an event
type PagerdutyAction string

const (
	Trigger                PagerdutyAction = "trigger"
	PagerdutyAckAction     PagerdutyAction = "acknowledge"
	PagerdutyResolveAction PagerdutyAction = "resolve"
)

// PagerdutySeverity represents the severity of an event
type PagerdutySeverity string

const (
	Critical PagerdutySeverity = "critical"
	Error    PagerdutySeverity = "error"
	Warning  PagerdutySeverity = "warning"
	Info     PagerdutySeverity = "info"
)

// PagerdutyResponseStatus is the response status of a Pagerduty API call
type PagerdutyResponseStatus string

const (
	SuccessStatus PagerdutyResponseStatus = "success"
)

type PagerdutyConfig struct {
	IntegrationKey  string
	ChangeEventsURL string
	AlertEventsURL  string
}

type PagerdutyClient interface {
	PostEvent(ctx context.Context, event *PagerdutyEventTrigger) (*PagerdutyAPIResponse, error)
}

type pagerdutyClient struct {
	integrationKey  string
	changeEventsURL string
	alertEventsURL  string
	client          *http.Client
}

// NewPagerdutyClient ... Initializer for Pagerduty client
func NewPagerdutyClient(cfg *PagerdutyConfig) PagerdutyClient {
	if cfg.IntegrationKey == "" {
		logging.NoContext().Warn("No Pagerduty integration key provided")
	}

	return pagerdutyClient{
		integrationKey:  cfg.IntegrationKey,
		changeEventsURL: cfg.ChangeEventsURL,
		alertEventsURL:  cfg.AlertEventsURL,
		client:          &http.Client{},
	}
}

// PagerdutyEventTrigger ... Represents caller specified fields for a Pagerduty event
type PagerdutyEventTrigger struct {
	Message  string
	Action   PagerdutyAction
	Severity PagerdutySeverity
	DedupKey string
}

// PagerdutyRequest ... Used to construct a Pagerduty api request
type PagerdutyRequest struct {
	RoutingKey  string           `json:"routing_key"`
	EventAction PagerdutyAction  `json:"event_action"`
	DedupKey    string           `json:"dedup_key"`
	Payload     PagerdutyPayload `json:"payload"`
}

// PagerdutyPayload ... Represents the payload of a Pagerduty event
type PagerdutyPayload struct {
	Summary   string            `json:"summary"`
	Source    string            `json:"source"`
	Severity  PagerdutySeverity `json:"severity"`
	Timestamp time.Time         `json:"timestamp"`
}

// newPagerdutyPayload ... Initializes a new Pagerduty payload given the integration key and event
func newPagerdutyPayload(integrationKey string, event *PagerdutyEventTrigger) *PagerdutyRequest {
	return &PagerdutyRequest{
		RoutingKey:  integrationKey,
		EventAction: event.Action,
		DedupKey:    event.DedupKey,
		Payload: PagerdutyPayload{
			Summary:   event.Message,
			Source:    "Pessimism",
			Severity:  event.Severity,
			Timestamp: time.Now(),
		},
	}
}

// marshal ... Marshals the Pagerduty payload
func (pdp *PagerdutyRequest) marshal() ([]byte, error) {
	bytes, err := json.Marshal(pdp)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// PagerdutyAPIResponse ... Represents the structure of a Pagerduty API response
type PagerdutyAPIResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	DedupKey string `json:"dedup_key"`
}

// PostEvent ... Posts a new event to Pagerduty
func (pdc pagerdutyClient) PostEvent(ctx context.Context, event *PagerdutyEventTrigger) (*PagerdutyAPIResponse, error) {
	// 1. Create and marshal payload into request object body

	payload, err := newPagerdutyPayload(pdc.integrationKey, event).marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pdc.alertEventsURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 2. Make request to Pagerduty
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

	var apiResp *PagerdutyAPIResponse
	if err := json.Unmarshal(bytes, &apiResp); err != nil {
		return nil, err
	}

	return apiResp, nil
}
