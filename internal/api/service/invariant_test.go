package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	testErrMsg1 = "69"
	testErrMsg2 = "420"
	testErrMsg3 = "666"
)

type testSuite struct {
	testCfg Config

	mockEngineMan *mocks.EngineManager
	mockEtlMan    *mocks.EtlManager

	apiSvc   Service
	mockCtrl *gomock.Controller
}

func testErr1() error {
	return fmt.Errorf(testErrMsg1)
}
func testErr2() error {
	return fmt.Errorf(testErrMsg2)
}
func testErr3() error {
	return fmt.Errorf(testErrMsg3)
}
func createTestSuite(ctrl *gomock.Controller, cfg Config) testSuite {
	engineManager := mocks.NewEngineManager(ctrl)
	etlManager := mocks.NewEtlManager(ctrl)

	svc := New(context.Background(), &cfg, etlManager, engineManager)
	return testSuite{
		testCfg: cfg,

		mockEngineMan: engineManager,
		mockEtlMan:    etlManager,

		apiSvc:   svc,
		mockCtrl: ctrl,
	}
}

func Test_ProcessInvariantRequest(t *testing.T) {
	ctrl := gomock.NewController(t)

	defaultRequestBody := func() models.InvRequestBody {
		return models.InvRequestBody{
			Method: models.Run,

			Params: models.InvRequestParams{
				Network: core.Layer1,
				PType:   core.Live,
				InvType: core.ExampleInv,

				StartHeight: nil,
				EndHeight:   nil,

				SessionParams: nil,
			},
		}
	}

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
				cfg := Config{}

				return createTestSuite(ctrl, cfg)
			},

			testLogic: func(t *testing.T, ts testSuite) {

				testParams := defaultRequestBody()
				testParams.Params.InvType = 42

				_, err := ts.apiSvc.ProcessInvariantRequest(testParams)

				assert.Error(t, err)
				assert.Contains(t, err.Error(), "could not find implementation for type")

			},
		},
		{
			name:        "Create Pipeline Failure",
			description: "When ProcessInvariantRequest is called that results in etl failure, an error should be returned",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPipelineUUID(), testErr1()).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {

				_, err := ts.apiSvc.ProcessInvariantRequest(defaultRequestBody())

				assert.Error(t, err)
				assert.Equal(t, err.Error(), testErr1().Error())

			},
		},
		{
			name:        "Deploy to Risk Engine Failure",
			description: "When ProcessInvariantRequest is called that results in a risk engine deploy failure, an error should be returned",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPipelineUUID(), nil).
					Times(1)

				ts.mockEngineMan.EXPECT().
					DeployInvariantSession(gomock.Any(), gomock.Any(),
						gomock.Any(), gomock.Any(), gomock.Any()).
					Return(core.NilInvariantUUID(), testErr2()).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {

				_, err := ts.apiSvc.ProcessInvariantRequest(defaultRequestBody())

				assert.Error(t, err)
				assert.Equal(t, err.Error(), testErr2().Error())

			},
		},
		{
			name:        "Run ETL Pipeline Failure",
			description: "When ProcessInvariantRequest is called that results in a pipeline that fails to run, an error should be returned",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPipelineUUID(), nil).
					Times(1)

				ts.mockEngineMan.EXPECT().
					DeployInvariantSession(gomock.Any(), gomock.Any(),
						gomock.Any(), gomock.Any(), gomock.Any()).
					Return(core.NilInvariantUUID(), nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					RunPipeline(core.NilPipelineUUID()).
					Return(testErr3())

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {

				_, err := ts.apiSvc.ProcessInvariantRequest(defaultRequestBody())

				assert.Error(t, err)
				assert.Equal(t, err.Error(), testErr3().Error())

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
