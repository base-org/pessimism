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

	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

type SlackConfig struct {
	Channel string
	URL     string
}

// SlackClient ... Interface for slack client
type SlackClient interface {
	PostData(context.Context, string) (*SlackAPIResponse, error)
}

// slackClient ... Slack client
type slackClient struct {
	url     string
	channel string
	client  *http.Client
}

// NewSlackClient ... Initializer
func NewSlackClient(cfg *SlackConfig) SlackClient {
	if cfg.URL == "" {
		logging.NoContext().Warn("No Slack webhook URL not provided")
	}

	return &slackClient{
		url: cfg.URL,
		// NOTE - This is a default client, we can add more configuration to it
		// when necessary
		channel: cfg.Channel,
		client:  &http.Client{},
	}
}

// slackPayload represents the structure of a slack alert
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
	Ok  bool   `json:"ok"`
	Err string `json:"error"`
}

// PostAlert ... handles posting data to slack
func (sc slackClient) PostData(ctx context.Context, str string) (*SlackAPIResponse, error) {
	// 1. make & marshal payload into request object body
	payload, err := newSlackPayload(str, sc.channel).marshal()
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

	return apiResp, err
}
