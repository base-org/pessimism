package alert

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// slackPayload represents the structure of a slack alert
type slackPayload struct {
	Text interface{} `json:"text"`
}

// newSlackPayload takes in an assessment, wraps it with Pagerduty payload body
// finally marshalls it
func newSlackPayload(body string) ([]byte, error) {
	p := &slackPayload{
		Text: body,
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// SlackHandler handlers processing an slack alert event
func SlackHandler(report string, client http.Client, url string) ([]byte, error) {
	// if the url is empty error
	if url == "" {
		return nil, errors.New("Slack webhook url not found. Check that a bad url wasn't passed with the event")
	}

	// make payload
	payload, err := newSlackPayload(report)
	if err != nil {
		return nil, errors.WithMessagef(err, "error generating slack payload")
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, errors.WithMessage(err, "error creating new http request for posting report to slack")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithMessage(err, "error posting report to slack")
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
