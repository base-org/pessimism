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

// SlackClient ... Interface for slack client
type SlackClient interface {
	PostData(context.Context, string) (*SlackAPIResponse, error)
}

// slackClient ... Slack client
type slackClient struct {
	url    string
	client *http.Client
}

// NewSlackClient ... Initializer
func NewSlackClient(url string) SlackClient {
	if url == "" {
		logging.NoContext().Warn("No Slack webhook URL not provided")
	}

	return slackClient{
		url: url,
		// NOTE - This is a default client, we can add more configuration to it
		// when necessary
		client: &http.Client{},
	}
}

// slackPayload represents the structure of a slack alert
type SlackPayload struct {
	Text interface{} `json:"text"`
}

// newSlackPayload ... initializes a new slack payload
func newSlackPayload(text interface{}) *SlackPayload {
	return &SlackPayload{Text: text}
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
	// make & marshal payload
	payload, err := newSlackPayload(str).marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost, sc.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// make request
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

	// read response
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
