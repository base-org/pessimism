package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/base-org/pessimism/internal/client"
)

// TestServer ... Mock server for testing slack alerts
type TestServer struct {
	Server   *httptest.Server
	Payloads []*client.SlackPayload
}

// NewTestServer ... Creates a new mock slack server
func NewTestServer() *TestServer {
	ts := &TestServer{
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
func (svr *TestServer) Close() {
	svr.Server.Close()
}

// mockSlackPost ... Mocks a slack post request
func (svr *TestServer) mockSlackPost(w http.ResponseWriter, r *http.Request) {
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
func (svr *TestServer) SlackAlerts() []*client.SlackPayload {
	return svr.Payloads
}

// ClearAlerts ... Clears the alerts
func (svr *TestServer) ClearAlerts() {
	svr.Payloads = []*client.SlackPayload{}
}
