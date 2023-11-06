package service_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func testErr() error {
	return fmt.Errorf("test error")
}

func Test_RunHeuristicSession(t *testing.T) {
	id := core.UUID{}

	ctrl := gomock.NewController(t)

	testCfg := &heuristic.DeployConfig{}

	defaultBody := &models.SessionRequestBody{
		Method: "run",
		Params: models.SessionRequestParams{
			Network:       "layer1",
			HeuristicType: "contract_event",
			StartHeight:   nil,
			EndHeight:     nil,
			SessionParams: nil,
		},
	}

	var tests = []struct {
		name string

		constructionLogic func() *testSuite
		testLogic         func(*testing.T, *testSuite)
	}{
		{
			name: "Successful heuristic session deployment",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPathCfg(&defaultBody.Params).
					Return(nil, nil).
					Times(1)

				ts.mockSub.EXPECT().
					BuildDeployCfg(gomock.Any(), gomock.Any()).
					Return(testCfg, nil).
					Times(1)

				ts.mockSub.EXPECT().
					RunHeuristic(testCfg).
					Return(id, nil).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := defaultBody.Clone()
				id, err := ts.apiSvc.ProcessHeuristicRequest(testParams)

				assert.NoError(t, err)
				assert.Equal(t, id, id)
			},
		},
		{
			name: "Failure when building path config",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPathCfg(&defaultBody.Params).
					Return(nil, testErr()).
					Times(1)
				return ts
			},

			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := defaultBody.Clone()
				id, err := ts.apiSvc.ProcessHeuristicRequest(testParams)

				assert.Error(t, err)
				assert.Equal(t, core.UUID{}, id)
			},
		},
		{
			name: "Failure when building deploy config",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPathCfg(&defaultBody.Params).
					Return(nil, nil).
					Times(1)

				ts.mockSub.EXPECT().
					BuildDeployCfg(gomock.Any(), gomock.Any()).
					Return(nil, testErr()).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := defaultBody.Clone()
				id, err := ts.apiSvc.ProcessHeuristicRequest(testParams)

				assert.Error(t, err)
				assert.Equal(t, core.UUID{}, id)
			},
		},
		{
			name: "Failure when running heuristic session",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPathCfg(&defaultBody.Params).
					Return(nil, nil).
					Times(1)

				ts.mockSub.EXPECT().
					BuildDeployCfg(gomock.Any(), gomock.Any()).
					Return(testCfg, nil).
					Times(1)

				ts.mockSub.EXPECT().
					RunHeuristic(testCfg).
					Return(core.UUID{}, testErr()).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := defaultBody.Clone()
				id, err := ts.apiSvc.ProcessHeuristicRequest(testParams)

				assert.Error(t, err)
				assert.Equal(t, core.UUID{}, id)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			testMeta := tc.constructionLogic()
			tc.testLogic(t, testMeta)
		})

	}

}
