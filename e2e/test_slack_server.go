package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/base-org/pessimism/internal/client"
)

// TestSlackServer ... Mock server for testing slack alerts
type TestSlackServer struct {
	Server   *httptest.Server
	Payloads []*client.SlackPayload
}

// NewTestSlackServer ... Creates a new mock slack server
func NewTestSlackServer() *TestSlackServer {
	ts := &TestSlackServer{
		Payloads: []*client.SlackPayload{},
	}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/":
			ts.mockSlackPost(w, r)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))

	return ts
}

// Close ... Closes the server
func (svr *TestSlackServer) Close() {
	svr.Server.Close()
}

// mockSlackPost ... Mocks a slack post request
func (svr *TestSlackServer) mockSlackPost(w http.ResponseWriter, r *http.Request) {
	var alert *client.SlackPayload

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false, "error":"could not decode slack payload"}`))
		return
	}

	svr.Payloads = append(svr.Payloads, alert)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true, "error":""}`))
}

// SlackAlerts ... Returns the slack alerts
func (svr *TestSlackServer) SlackAlerts() []*client.SlackPayload {
	return svr.Payloads
}

// ClearAlerts ... Clears the alerts
func (svr *TestSlackServer) ClearAlerts() {
	svr.Payloads = []*client.SlackPayload{}
}
