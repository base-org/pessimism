package subsystem_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/base-org/pessimism/internal/subsystem"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func testErr() error {
	return fmt.Errorf("test error")
}

type testSuite struct {
	subsys subsystem.Manager

	mockEtl  *mocks.EtlManager
	mockEng  *mocks.EngineManager
	mockAlrt *mocks.AlertManager
	mockCtrl *gomock.Controller
}

func createTestSuite(t *testing.T) *testSuite {
	ctrl := gomock.NewController(t)

	etlMock := mocks.NewEtlManager(ctrl)
	engMock := mocks.NewEngineManager(ctrl)
	alrtMock := mocks.NewAlertManager(ctrl)
	cfg := &subsystem.Config{}

	subsys := subsystem.NewManager(context.Background(), cfg, etlMock, engMock, alrtMock)

	return &testSuite{
		subsys:   subsys,
		mockEtl:  etlMock,
		mockEng:  engMock,
		mockAlrt: alrtMock,
		mockCtrl: ctrl,
	}
}

func Test_BuildDeployCfg(t *testing.T) {
	pConfig := &core.PipelineConfig{
		Network:      core.Layer1,
		DataType:     core.GethBlock,
		PipelineType: core.Live,
		ClientConfig: nil,
	}

	sConfig := &core.SessionConfig{
		Network: core.Layer1,
		PT:      core.Live,
		Type:    core.BalanceEnforcement,
		Params:  nil,
	}

	var tests = []struct {
		name        string
		constructor func(t *testing.T) *testSuite
		testLogic   func(t *testing.T, ts *testSuite)
	}{
		{
			name: "Failure when fetching state key",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)

				ts.mockEtl.EXPECT().GetStateKey(pConfig.DataType).
					Return(nil, false, testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualCfg, err := ts.subsys.BuildDeployCfg(pConfig, sConfig)
				assert.Error(t, err)
				assert.Nil(t, actualCfg)
			},
		},
		{
			name: "Failure when creating data pipeline",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)

				ts.mockEtl.EXPECT().GetStateKey(pConfig.DataType).
					Return(nil, false, nil).
					Times(1)

				ts.mockEtl.EXPECT().CreateDataPipeline(gomock.Any()).
					Return(core.NilPUUID(), false, testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualCfg, err := ts.subsys.BuildDeployCfg(pConfig, sConfig)
				assert.Error(t, err)
				assert.Nil(t, actualCfg)
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, test.name), func(t *testing.T) {
			ts := test.constructor(t)
			test.testLogic(t, ts)
		})
	}
}

func Test_RunInvSession(t *testing.T) {
	testSUUID := core.MakeSUUID(1, 1, 1)
	testCfg := &invariant.DeployConfig{
		Stateful: false,
		StateKey: nil,
		Network:  core.Layer1,
		PUUID:    core.NilPUUID(),
		Reuse:    false,

		InvType:   core.BalanceEnforcement,
		InvParams: nil,
		AlertDest: core.Slack,
	}

	var tests = []struct {
		name        string
		constructor func(t *testing.T) *testSuite
		testLogic   func(t *testing.T, ts *testSuite)
	}{
		{
			name: "Failure when deploying invariant session",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockEng.EXPECT().DeployInvariantSession(testCfg).
					Return(core.NilSUUID(), testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualSUUID, err := ts.subsys.RunInvSession(testCfg)
				assert.Error(t, err)
				assert.Equal(t, core.NilSUUID(), actualSUUID)
			},
		},
		{
			name: "Failure when adding invariant session to alerting system",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockEng.EXPECT().DeployInvariantSession(testCfg).
					Return(testSUUID, nil).
					Times(1)

				ts.mockAlrt.EXPECT().AddSession(testSUUID, testCfg.AlertDest).
					Return(testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualSUUID, err := ts.subsys.RunInvSession(testCfg)
				assert.Error(t, err)
				assert.Equal(t, core.NilSUUID(), actualSUUID)
			},
		},
		{
			name: "Success with no reuse",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockEng.EXPECT().DeployInvariantSession(testCfg).
					Return(testSUUID, nil).
					Times(1)

				ts.mockAlrt.EXPECT().AddSession(testSUUID, testCfg.AlertDest).
					Return(nil).
					Times(1)

				ts.mockEtl.EXPECT().RunPipeline(testCfg.PUUID).
					Return(nil).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualSUUID, err := ts.subsys.RunInvSession(testCfg)
				assert.NoError(t, err)
				assert.Equal(t, testSUUID, actualSUUID)
			},
		},
		{
			name: "Success with reuse",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockEng.EXPECT().DeployInvariantSession(testCfg).
					Return(testSUUID, nil).
					Times(1)

				ts.mockAlrt.EXPECT().AddSession(testSUUID, testCfg.AlertDest).
					Return(nil).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				testCfg.Reuse = true
				actualSUUID, err := ts.subsys.RunInvSession(testCfg)
				assert.NoError(t, err)
				assert.Equal(t, testSUUID, actualSUUID)
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, test.name), func(t *testing.T) {
			ts := test.constructor(t)
			test.testLogic(t, ts)
		})
	}
}
