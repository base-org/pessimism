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
	Server      *httptest.Server
	SlackAlerts []*client.SlackPayload
}

// MockSlackServer ... Creates a new mock slack server
func MockSlackServer() *TestServer {
	ts := &TestServer{
		SlackAlerts: []*client.SlackPayload{},
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

// mockSlackPost ... Mocks a slack post request
func (svr *TestServer) mockSlackPost(w http.ResponseWriter, r *http.Request) {
	var alert *client.SlackPayload

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false, "error":"could not decode slack payload"}`))
		return
	}

	svr.SlackAlerts = append(svr.SlackAlerts, alert)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true, "error":""}`))
}
