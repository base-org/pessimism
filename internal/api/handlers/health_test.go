package handlers_test

import (
	"fmt"
	"io/ioutil"
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
			name:        "Get Invariant Failure",
			description: "When ProcessInvariantRequest is called provided an invalid invariant, an error should be returned",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				ts := createTestSuite(t)
				ts.mockSvc.EXPECT().
					CheckHealth().
					Return(&models.HealthCheck{Healthy: true}).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, testAddress, nil)

				ts.testHandler.HealthCheck(w, r)
				res := w.Result()

				data, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Errorf("Error: %v", err)
				}

				actualHc := models.HealthCheck{}
				err = actualHc.UnmarshalJson(data)

				assert.NoError(t, err)
				assert.True(t, actualHc.Healthy)
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
