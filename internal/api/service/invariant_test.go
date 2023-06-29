package service_test

import (
	"fmt"
	"testing"

	svc "github.com/base-org/pessimism/internal/api/service"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_ProcessInvariantRequest(t *testing.T) {
	ctrl := gomock.NewController(t)

	defaultRequestBody := func() models.InvRequestBody {
		return models.InvRequestBody{
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
				cfg := svc.Config{}

				return createTestSuite(ctrl, cfg)
			},

			testLogic: func(t *testing.T, ts testSuite) {

				testParams := defaultRequestBody()
				testParams.Params.InvType = "bleh"

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
				cfg := svc.Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPUUID(), false, testErr1()).
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
			description: "When ProcessInvariantRequest is called that results in a etl registry fetch failure, an error should be returned",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := svc.Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPUUID(), false, nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					GetRegister(gomock.Any()).
					Return(nil, testErr1()).
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
				cfg := svc.Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPUUID(), false, nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					GetRegister(gomock.Any()).
					Return(&core.DataRegister{}, nil).
					Times(1)

				ts.mockEngineMan.EXPECT().
					DeployInvariantSession(gomock.Any()).
					Return(core.NilSUUID(), testErr2()).
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
				cfg := svc.Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockAlertMan.EXPECT().
					AddInvariantSession(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPUUID(), false, nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					GetRegister(gomock.Any()).
					Return(&core.DataRegister{}, nil).
					Times(1)

				ts.mockEngineMan.EXPECT().
					DeployInvariantSession(gomock.Any()).
					Return(core.NilSUUID(), nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					RunPipeline(core.NilPUUID()).
					Return(testErr3()).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {

				_, err := ts.apiSvc.ProcessInvariantRequest(defaultRequestBody())

				assert.Error(t, err)
				assert.Equal(t, err.Error(), testErr3().Error())

			},
		},
		{
			name:        "Successful Sesion Creation",
			description: "When ProcessInvariantRequest is called that results in a pipeline that succeeds to run, an invariant UUID should be returned",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := svc.Config{}

				ts := createTestSuite(ctrl, cfg)

				ts.mockEtlMan.EXPECT().
					CreateDataPipeline(gomock.Any()).
					Return(core.NilPUUID(), false, nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					GetRegister(gomock.Any()).
					Return(&core.DataRegister{}, nil).
					Times(1)

				ts.mockEngineMan.EXPECT().
					DeployInvariantSession(gomock.Any()).
					Return(testSUUID1(), nil).
					Times(1)

				ts.mockEtlMan.EXPECT().
					RunPipeline(core.NilPUUID()).
					Return(nil).
					Times(1)

				ts.mockAlertMan.EXPECT().
					AddInvariantSession(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {

				sUUUID, err := ts.apiSvc.ProcessInvariantRequest(defaultRequestBody())

				assert.NoError(t, err)
				assert.Equal(t, testSUUID1().PID.String(), sUUUID.PID.String())

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
