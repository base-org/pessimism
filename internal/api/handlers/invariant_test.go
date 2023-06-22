package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_ProcessInvariantRequest(t *testing.T) {

	var tests = []struct {
		name        string
		description string
		function    string

		constructionLogic func() testSuite
		testLogic         func(*testing.T, testSuite)
	}{
		{
			name:        "Get Invariant Failure",
			description: "When provided a malformed request body, a failed decoding response should be returned",
			function:    "RunInvariant",

			constructionLogic: func() testSuite {
				ts := createTestSuite(t)
				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				w := httptest.NewRecorder()

				testBody := bytes.NewBuffer([]byte{0x42})
				r := httptest.NewRequest(http.MethodGet, testAddress, testBody)

				ts.testHandler.RunInvariant(w, r)
				res := w.Result()

				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("Error: %v", err)
				}

				actualResp := &models.InvResponse{}
				err = json.Unmarshal(data, actualResp)

				assert.NoError(t, err)
				assert.Equal(t, models.NewInvUnmarshalErrResp(), actualResp)
			},
		},
		{
			name:        "Process Invariant Failure",
			description: "When provided an internal error occurs, a failed processing response should be returned",
			function:    "RunInvariant",

			constructionLogic: func() testSuite {
				ts := createTestSuite(t)

				ts.mockSvc.EXPECT().
					ProcessInvariantRequest(gomock.Any()).
					Return(core.NilSUUID(), testError1()).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				w := httptest.NewRecorder()

				testBody, _ := json.Marshal(models.InvRequestBody{Method: "run"})

				testBytes := bytes.NewBuffer(testBody)
				r := httptest.NewRequest(http.MethodGet, testAddress, testBytes)

				ts.testHandler.RunInvariant(w, r)
				res := w.Result()

				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("Error: %v", err)
				}

				actualResp := &models.InvResponse{}
				err = json.Unmarshal(data, actualResp)

				assert.NoError(t, err)
				assert.Equal(t, models.NewInvNoProcessInvResp(), actualResp)
			},
		},
		{
			name:        "Process Invariant Success",
			description: "When an invariant is successfully processed, a suuid should be rendered",
			function:    "RunInvariant",

			constructionLogic: func() testSuite {
				ts := createTestSuite(t)

				ts.mockSvc.EXPECT().
					ProcessInvariantRequest(gomock.Any()).
					Return(testSUUID1(), nil).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				w := httptest.NewRecorder()

				testBody, _ := json.Marshal(models.InvRequestBody{Method: "run"})

				testBytes := bytes.NewBuffer(testBody)
				r := httptest.NewRequest(http.MethodGet, testAddress, testBytes)

				ts.testHandler.RunInvariant(w, r)
				res := w.Result()

				data, err := io.ReadAll(res.Body)
				if err != nil {
					t.Errorf("Error: %v", err)
				}

				actualResp := &models.InvResponse{}
				err = json.Unmarshal(data, actualResp)

				assert.NoError(t, err)

				assert.Equal(t, actualResp.Status, models.OK)
				assert.Equal(t, actualResp.Code, http.StatusAccepted)
				assert.Contains(t, actualResp.Result[core.SUUIDKey], testSUUID1().PID.String())
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
