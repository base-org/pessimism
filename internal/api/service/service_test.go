package service_test

import (
	"context"
	"fmt"

	svc "github.com/base-org/pessimism/internal/api/service"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
)

const (
	testErrMsg1 = "69"
	testErrMsg2 = "420"
	testErrMsg3 = "666"
)

type testSuite struct {
	testCfg svc.Config

	mockAlertMan  *mocks.MockAlertingManager
	mockEngineMan *mocks.EngineManager
	mockEtlMan    *mocks.EtlManager

	apiSvc   svc.Service
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

func createTestSuite(ctrl *gomock.Controller, cfg svc.Config) testSuite {
	engineManager := mocks.NewEngineManager(ctrl)
	etlManager := mocks.NewEtlManager(ctrl)
	alertManager := mocks.NewMockAlertingManager(ctrl)

	service := svc.New(context.Background(), &cfg, alertManager, etlManager, engineManager)
	return testSuite{
		testCfg: cfg,

		mockAlertMan:  alertManager,
		mockEngineMan: engineManager,
		mockEtlMan:    etlManager,

		apiSvc:   service,
		mockCtrl: ctrl,
	}
}
