//go:generate mockgen -package mocks --destination ../mocks/slack_client.go . SlackClient

package client

// NOTE - API endpoint specifications for slack client
// can be found here - https://api.slack.com/methods/chat.postMessage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

type SlackClient interface {
	AlertClient
}

type SlackConfig struct {
	Channel  string
	URL      string
	Priority string
}

// slackClient ... Slack client
type slackClient struct {
	name    string
	url     string
	channel string
	client  *http.Client
}

// NewSlackClient ... Initializer
func NewSlackClient(cfg *SlackConfig, name string) SlackClient {
	if cfg.URL == "" {
		logging.NoContext().Warn("No Slack webhook URL not provided")
	}

	return &slackClient{
		url:  cfg.URL,
		name: name,
		// NOTE - This is a default client, we can add more configuration to it
		// when necessary
		channel: cfg.Channel,
		client:  &http.Client{},
	}
}

// SlackPayload represents the structure of a slack alert
type SlackPayload struct {
	Text    interface{} `json:"text"`
	Channel string      `json:"channel"`
}

// newSlackPayload ... initializes a new slack payload
func newSlackPayload(text interface{}, channel string) *SlackPayload {
	return &SlackPayload{Text: text, Channel: channel}
}

// marshal ... marshals the slack payload
func (sp *SlackPayload) marshal() ([]byte, error) {
	bytes, err := json.Marshal(sp)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// SlackAPIResponse ... represents the structure of a slack API response
type SlackAPIResponse struct {
	Message string `json:"message"`
	Err     string `json:"error"`
}

// ToAlertResponse ... Converts a slack API response to an alert API response
func (a *SlackAPIResponse) ToAlertResponse() *AlertAPIResponse {
	status := core.SuccessStatus
	if a.Message != "ok" {
		status = core.FailureStatus
	}

	return &AlertAPIResponse{
		Status:  status,
		Message: a.Err,
	}
}

// PostEvent ... handles posting an event to slack
func (sc slackClient) PostEvent(ctx context.Context, event *AlertEventTrigger) (*AlertAPIResponse, error) {
	// 1. make & marshal payload into request object body
	payload, err := newSlackPayload(event.Message, sc.channel).marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, sc.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 2. make request to slack
	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logging.WithContext(ctx).Warn("Could not close slack response body",
				zap.Error(err))
		}
	}()

	// 3. read and unmarshal response
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp *SlackAPIResponse
	if err := json.Unmarshal(bytes, &apiResp); err != nil {
		return nil, err
	}

	return apiResp.ToAlertResponse(), nil
}

// GetName ... returns the name of the slack client
func (sc slackClient) GetName() string {
	return sc.name
}
