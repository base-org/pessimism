package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/logging"

	"go.uber.org/zap"
)

// TestPagerdutyServer ... Mock server for testing pagerduty alerts
type TestPagerdutyServer struct {
	Server   *httptest.Server
	Payloads []*client.PagerdutyRequest
}

// NewTestPagerdutyServer ... Creates a new mock pagerduty server
func NewTestPagerdutyServer() *TestPagerdutyServer {
	ts := &TestPagerdutyServer{
		Payloads: []*client.PagerdutyRequest{},
	}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/":
			ts.mockPagerdutyPost(w, r)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))

	return ts
}

// Close ... Closes the server
func (svr *TestPagerdutyServer) Close() {
	svr.Server.Close()
}

// mockPagerdutyPost ... Mocks a pagerduty post request
func (svr *TestPagerdutyServer) mockPagerdutyPost(w http.ResponseWriter, r *http.Request) {
	var alert *client.PagerdutyRequest

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":false, "message":"could not decode pagerduty payload"}`))
		return
	}

	svr.Payloads = append(svr.Payloads, alert)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":success, "message":""}`))
}

// PagerdutyAlerts ... Returns the pagerduty alerts
func (svr *TestPagerdutyServer) PagerdutyAlerts() []*client.PagerdutyRequest {
	logging.NoContext().Info("Payloads", zap.Any("payloads", svr.Payloads))

	return svr.Payloads
}

// ClearAlerts ... Clears the alerts
func (svr *TestPagerdutyServer) ClearAlerts() {
	svr.Payloads = []*client.PagerdutyRequest{}
}
