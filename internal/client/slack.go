package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/base-org/pessimism/internal/logging"
)

// SlackClient ... Interface for slack client
type SlackClient interface {
	PostAlert(report string) ([]byte, error)
}

// slackClient ... Slack client
type slackClient struct {
	url    string
	client *http.Client
}

// NewSlackClient ... Initializer
func NewSlackClient(url string) slackClient {

	if url == "" {
		logging.NoContext().Warn("Slack URL not provided")
	}

	return slackClient{
		url: url,
		// NOTE - This is a default client, we can add more configuration to it
		// when necessary
		client: &http.Client{},
	}
}

// slackPayload represents the structure of a slack alert
type slackPayload struct {
	Text interface{} `json:"text"`
}

// newSlackPayload ... initializes a new slack payload
func newSlackPayload(text interface{}) *slackPayload {
	return &slackPayload{Text: text}
}

// marshal ... marshals the slack payload
func (sp *slackPayload) marshal() ([]byte, error) {
	bytes, err := json.Marshal(sp)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// PostAlert handlers processing a slack alert event
func (sc slackClient) PostAlert(report string) ([]byte, error) {

	// make & marshal payload
	payload, err := newSlackPayload(report).marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, sc.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// make request
	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
