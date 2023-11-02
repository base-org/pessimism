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
	Port     int
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

	err = pds.Server.Listener.Close()
	if err != nil {
		panic(err)
	}
	pds.Server.Listener = l

	// get port from listener
	pds.Port = pds.Server.Listener.Addr().(*net.TCPAddr).Port
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

// TestSlackServer ... Mock server for testing slack alerts
type TestSlackServer struct {
	Server       *httptest.Server
	Payloads     []*client.SlackPayload
	Port         int
	Unstructured bool
}

// NewTestSlackServer ... Creates a new mock slack server
func NewTestSlackServer(url string, port int) *TestSlackServer { //nolint:dupl //This will be addressed
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", url, port))
	if err != nil {
		panic(err)
	}

	ss := &TestSlackServer{
		Payloads: []*client.SlackPayload{},
	}

	ss.Server = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimSpace(r.URL.Path) {
		case "/":
			ss.mockSlackPost(w, r)
		default:
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	}))

	err = ss.Server.Listener.Close()
	if err != nil {
		panic(err)
	}
	ss.Server.Listener = l
	// get port from listener
	ss.Port = ss.Server.Listener.Addr().(*net.TCPAddr).Port

	ss.Server.Start()

	logging.NoContext().Info("Test slack server started", zap.String("url", url), zap.Int("port", port))

	return ss
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
		_, _ = w.Write([]byte(`{"message":"", "error":"could not decode slack payload"}`))
		return
	}

	svr.Payloads = append(svr.Payloads, alert)
	w.WriteHeader(http.StatusOK)
	if svr.Unstructured {
		_, _ = w.Write([]byte(`ok`))
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok", "error":""}`))
	}
}

// SlackAlerts ... Returns the slack alerts
func (svr *TestSlackServer) SlackAlerts() []*client.SlackPayload {
	return svr.Payloads
}

// ClearAlerts ... Clears the alerts
func (svr *TestSlackServer) ClearAlerts() {
	svr.Payloads = []*client.SlackPayload{}
}
