package service

import (
	"context"
	"fmt"
	"testing"

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

func testSUUID1() core.InvSessionUUID {
	return core.MakeInvSessionUUID(1, 1, 1)
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

func Test_GetHealth(t *testing.T) {
	ctrl := gomock.NewController(t)

	var tests = []struct {
		name        string
		description string
		function    string

		constructionLogic func() testSuite
		testLogic         func(*testing.T, testSuite)
	}{
		{
			name:        "Get Health Success",
			description: "",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := Config{}

				return createTestSuite(ctrl, cfg)
			},

			testLogic: func(t *testing.T, ts testSuite) {
				hc := ts.apiSvc.CheckHealth()

				assert.True(t, hc.Healthy)
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
