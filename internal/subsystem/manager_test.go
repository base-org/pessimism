package subsystem_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/base-org/pessimism/internal/subsystem"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func testErr() error {
	return fmt.Errorf("test error")
}

type testSuite struct {
	sys *subsystem.Manager

	mockETL   *mocks.MockETL
	mockENG   *mocks.EngineManager
	mockAlert *mocks.AlertManager
	mockCtrl  *gomock.Controller
}

func createTestSuite(t *testing.T) *testSuite {
	ctrl := gomock.NewController(t)

	etlMock := mocks.NewMockETL(ctrl)
	engMock := mocks.NewEngineManager(ctrl)
	alrtMock := mocks.NewAlertManager(ctrl)
	cfg := &subsystem.Config{
		MaxPathCount: 10,
	}

	sys := subsystem.NewManager(context.Background(), cfg, etlMock, engMock, alrtMock)

	return &testSuite{
		sys:       sys,
		mockETL:   etlMock,
		mockENG:   engMock,
		mockAlert: alrtMock,
		mockCtrl:  ctrl,
	}
}

func TestBuildDeployCfg(t *testing.T) {
	pConfig := &core.PathConfig{
		Network:      core.Layer1,
		DataType:     core.BlockHeader,
		PathType:     core.Live,
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

				ts.mockETL.EXPECT().GetStateKey(pConfig.DataType).
					Return(nil, false, testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualCfg, err := ts.sys.BuildDeployCfg(pConfig, sConfig)
				assert.Error(t, err)
				assert.Nil(t, actualCfg)
			},
		},
		{
			name: "Failure when creating data path",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)

				ts.mockETL.EXPECT().GetStateKey(pConfig.DataType).
					Return(nil, false, nil).
					Times(1)

				ts.mockETL.EXPECT().CreateProcessPath(gomock.Any()).
					Return(core.PathID{}, false, testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualCfg, err := ts.sys.BuildDeployCfg(pConfig, sConfig)
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

func TestRunHeuristic(t *testing.T) {
	id := core.NewUUID()
	testCfg := &heuristic.DeployConfig{
		Stateful: false,
		StateKey: nil,
		Network:  core.Layer1,
		PathID:   core.PathID{},
		Reuse:    false,

		HeuristicType: core.BalanceEnforcement,
		Params:        nil,
		AlertingPolicy: &core.AlertPolicy{
			Dest: core.Slack.String(),
		}}

	var tests = []struct {
		name        string
		constructor func(t *testing.T) *testSuite
		testLogic   func(t *testing.T, ts *testSuite)
	}{
		{
			name: "Failure when deploying heuristic session",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockETL.EXPECT().
					ActiveCount().Return(1).
					Times(1)

				ts.mockENG.EXPECT().DeployHeuristic(testCfg).
					Return(core.UUID{}, testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualSUUID, err := ts.sys.RunHeuristic(testCfg)
				assert.Error(t, err)
				assert.Equal(t, core.UUID{}, actualSUUID)
			},
		},
		{
			name: "Failure when adding heuristic session to alerting system",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockETL.EXPECT().
					ActiveCount().Return(1).
					Times(1)

				ts.mockENG.EXPECT().DeployHeuristic(testCfg).
					Return(id, nil).
					Times(1)

				ts.mockAlert.EXPECT().AddSession(id, testCfg.AlertingPolicy).
					Return(testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualSUUID, err := ts.sys.RunHeuristic(testCfg)
				assert.Error(t, err)
				assert.Equal(t, core.UUID{}, actualSUUID)
			},
		},
		{
			name: "Success with no reuse",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)

				ts.mockETL.EXPECT().
					ActiveCount().Return(1).
					Times(1)

				ts.mockENG.EXPECT().DeployHeuristic(testCfg).
					Return(id, nil).
					Times(1)

				ts.mockAlert.EXPECT().AddSession(id, testCfg.AlertingPolicy).
					Return(nil).
					Times(1)

				ts.mockETL.EXPECT().Run(testCfg.PathID).
					Return(nil).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				actualSUUID, err := ts.sys.RunHeuristic(testCfg)
				assert.NoError(t, err)
				assert.Equal(t, id, actualSUUID)
			},
		},
		{
			name: "Success with reuse",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)

				ts.mockENG.EXPECT().DeployHeuristic(testCfg).
					Return(id, nil).
					Times(1)

				ts.mockAlert.EXPECT().AddSession(id, testCfg.AlertingPolicy).
					Return(nil).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				testCfg.Reuse = true
				actualSUUID, err := ts.sys.RunHeuristic(testCfg)
				assert.NoError(t, err)
				assert.Equal(t, id, actualSUUID)
			},
		},
		{
			name: "Failure when active path count is reached",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)

				ts.mockETL.EXPECT().
					ActiveCount().Return(10).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				testCfg.Reuse = false
				actualSUUID, err := ts.sys.RunHeuristic(testCfg)
				assert.Error(t, err)
				assert.Equal(t, core.UUID{}, actualSUUID)
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

func TestBuildPathCfg(t *testing.T) {

	var tests = []struct {
		name        string
		constructor func(t *testing.T) *testSuite
		testLogic   func(t *testing.T, ts *testSuite)
	}{
		{
			name: "Failure when getting input type",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockENG.EXPECT().GetInputType(core.BalanceEnforcement).
					Return(core.BlockHeader, testErr()).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := &models.SessionRequestParams{
					Network:       core.Layer1.String(),
					HeuristicType: core.BalanceEnforcement.String(),
				}

				cfg, err := ts.sys.BuildPathCfg(testParams)
				assert.Error(t, err)
				assert.Nil(t, cfg)
			},
		},
		{
			name: "Failure when getting poll interval for invalid network",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockENG.EXPECT().GetInputType(core.BalanceEnforcement).
					Return(core.BlockHeader, nil).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := &models.SessionRequestParams{
					Network:       "layer0",
					HeuristicType: core.BalanceEnforcement.String(),
				}

				cfg, err := ts.sys.BuildPathCfg(testParams)
				assert.Error(t, err)
				assert.Nil(t, cfg)
			},
		},
		{
			name: "Success with valid params",
			constructor: func(t *testing.T) *testSuite {
				ts := createTestSuite(t)
				ts.mockENG.EXPECT().GetInputType(core.BalanceEnforcement).
					Return(core.BlockHeader, nil).
					Times(1)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				testParams := &models.SessionRequestParams{
					Network:       core.Layer1.String(),
					HeuristicType: core.BalanceEnforcement.String(),
				}

				cfg, err := ts.sys.BuildPathCfg(testParams)
				assert.NoError(t, err)
				assert.NotNil(t, cfg)

				assert.Equal(t, core.Layer1, cfg.Network)
				assert.Equal(t, core.Live, cfg.PathType)
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
