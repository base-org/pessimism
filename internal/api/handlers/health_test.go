package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/stretchr/testify/assert"
)

const (
	testAddress = "http://abc.xyz"
)

func Test_HealthCheck(t *testing.T) {

	var tests = []struct {
		name        string
		description string
		function    string

		constructionLogic func() testSuite
		testLogic         func(*testing.T, testSuite)
	}{
		{
			name:        "Successful Health Check",
			description: "When GetHealth is called provided a healthy application, a healthy check should be rendered",
			function:    "GetHealth",

			constructionLogic: func() testSuite {
				ts := createTestSuite(t)
				ts.mockSvc.EXPECT().
					CheckHealth().
					Return(&models.HealthCheck{
						Healthy: true,
						ChainConnectionStatus: &models.ChainConnectionStatus{
							IsL1Healthy: true,
							IsL2Healthy: true,
						}}).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, testAddress, nil)

				ts.testHandler.HealthCheck(w, r)
				res := w.Result()

				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("Error: %v", err)
				}

				actualHc := &models.HealthCheck{}
				err = json.Unmarshal(data, actualHc)

				assert.NoError(t, err)
				assert.True(t, actualHc.Healthy)
				assert.True(t, actualHc.ChainConnectionStatus.IsL1Healthy)
				assert.True(t, actualHc.ChainConnectionStatus.IsL2Healthy)
			},
		},
		{
			name:        "Failed Health Check",
			description: "When GetHealth is called provided a unhealthy application, an unhealthy check should be rendered",
			function:    "GetHealth",

			constructionLogic: func() testSuite {
				ts := createTestSuite(t)
				ts.mockSvc.EXPECT().
					CheckHealth().
					Return(&models.HealthCheck{
						Healthy: false,
						ChainConnectionStatus: &models.ChainConnectionStatus{
							IsL1Healthy: false,
							IsL2Healthy: true,
						}}).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, testAddress, nil)

				ts.testHandler.HealthCheck(w, r)
				res := w.Result()

				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("Error: %v", err)
				}

				actualHc := &models.HealthCheck{}
				err = json.Unmarshal(data, actualHc)

				assert.NoError(t, err)
				assert.False(t, actualHc.Healthy)
				assert.False(t, actualHc.ChainConnectionStatus.IsL1Healthy)
				assert.True(t, actualHc.ChainConnectionStatus.IsL2Healthy)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.name, tc.function), func(t *testing.T) {
			testMeta := tc.constructionLogic()
			tc.testLogic(t, testMeta)
		})

	}

}
