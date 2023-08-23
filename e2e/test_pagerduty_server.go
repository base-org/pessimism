package e2e

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/logging"

	"go.uber.org/zap"
)

// TestPagerDutyServer ... Mock server for testing pagerduty alerts
type TestPagerDutyServer struct {
	Server   *httptest.Server
	Payloads []*client.PagerDutyRequest
}

// NewTestPagerDutyServer ... Creates a new mock pagerduty server
func NewTestPagerDutyServer(url string, port int) *TestPagerDutyServer { //nolint:dupl //This will be addressed
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", url, port))
	if err != nil {
		panic(err)
	}

	pds := &TestPagerDutyServer{
		Payloads: []*client.PagerDutyRequest{},
	}

	pds.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/":
			pds.mockPagerDutyPost(w, r)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))

	pds.Server.Listener.Close()
	pds.Server.Listener = l
	pds.Server.Start()

	logging.NoContext().Info("Test pagerduty server started", zap.String("url", url), zap.Int("port", port))

	return pds
}

// Close ... Closes the server
func (svr *TestPagerDutyServer) Close() {
	svr.Server.Close()
}

// mockPagerDutyPost ... Mocks a pagerduty post request
func (svr *TestPagerDutyServer) mockPagerDutyPost(w http.ResponseWriter, r *http.Request) {
	var alert *client.PagerDutyRequest

	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":"failure"", "message":"could not decode pagerduty payload"}`))
		return
	}

	svr.Payloads = append(svr.Payloads, alert)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success", "message":""}`))
}

// PagerDutyAlerts ... Returns the pagerduty alerts
func (svr *TestPagerDutyServer) PagerDutyAlerts() []*client.PagerDutyRequest {
	logging.NoContext().Info("Payloads", zap.Any("payloads", svr.Payloads))

	return svr.Payloads
}

// ClearAlerts ... Clears the alerts
func (svr *TestPagerDutyServer) ClearAlerts() {
	svr.Payloads = []*client.PagerDutyRequest{}
}
