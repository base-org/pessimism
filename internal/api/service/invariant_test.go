package service_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func testErr() error {
	return fmt.Errorf("test error")
}

func Test_RunInvariantSession(t *testing.T) {
	testSUUID := core.MakeSUUID(1, 1, 1)

	ctrl := gomock.NewController(t)

	testCfg := &invariant.DeployConfig{}

	defaultBody := &models.InvRequestBody{
		Method: "run",

		Params: models.InvRequestParams{
			Network: "layer1",
			PType:   "live",
			InvType: "contract_event",

			StartHeight: nil,
			EndHeight:   nil,

			SessionParams: nil,
		},
	}

	var tests = []struct {
		name string

		constructionLogic func() *testSuite
		testLogic         func(*testing.T, *testSuite)
	}{
		{
			name: "Successful invariant session deployment",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPipelineCfg(&defaultBody.Params).
					Return(nil, nil).
					Times(1)

				ts.mockSub.EXPECT().
					BuildDeployCfg(gomock.Any(), gomock.Any()).
					Return(testCfg, nil).
					Times(1)

				ts.mockSub.EXPECT().
					RunInvSession(testCfg).
					Return(testSUUID, nil).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts *testSuite) {

				testParams := defaultBody.Clone()
				actualSUUID, err := ts.apiSvc.ProcessInvariantRequest(testParams)

				assert.NoError(t, err)
				assert.Equal(t, testSUUID, actualSUUID)
			},
		},
		{
			name: "Failure when building pipeline config",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPipelineCfg(&defaultBody.Params).
					Return(nil, testErr()).
					Times(1)
				return ts
			},

			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := defaultBody.Clone()
				actualSUUID, err := ts.apiSvc.ProcessInvariantRequest(testParams)

				assert.Error(t, err)
				assert.Equal(t, core.NilSUUID(), actualSUUID)
			},
		},
		{
			name: "Failure when building deploy config",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPipelineCfg(&defaultBody.Params).
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
				actualSUUID, err := ts.apiSvc.ProcessInvariantRequest(testParams)

				assert.Error(t, err)
				assert.Equal(t, core.NilSUUID(), actualSUUID)
			},
		},
		{
			name: "Failure when running invariant session",
			constructionLogic: func() *testSuite {
				ts := createTestSuite(ctrl)

				ts.mockSub.EXPECT().
					BuildPipelineCfg(&defaultBody.Params).
					Return(nil, nil).
					Times(1)

				ts.mockSub.EXPECT().
					BuildDeployCfg(gomock.Any(), gomock.Any()).
					Return(testCfg, nil).
					Times(1)

				ts.mockSub.EXPECT().
					RunInvSession(testCfg).
					Return(core.NilSUUID(), testErr()).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := defaultBody.Clone()
				actualSUUID, err := ts.apiSvc.ProcessInvariantRequest(testParams)

				assert.Error(t, err)
				assert.Equal(t, core.NilSUUID(), actualSUUID)
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
