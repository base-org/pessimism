package service_test

import (
	"context"
	"fmt"

	svc "github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/subsystem"

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

	mockAlertMan           *mocks.AlertManager
	mockEngineMan          *mocks.EngineManager
	mockEtlMan             *mocks.EtlManager
	mockEthClientInterface *mocks.MockEthClientInterface

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

func testSUUID1() core.SUUID {
	return core.MakeSUUID(1, 1, 1)
}

func createTestSuite(ctrl *gomock.Controller, cfg svc.Config) testSuite {
	engineManager := mocks.NewEngineManager(ctrl)
	etlManager := mocks.NewEtlManager(ctrl)
	alertManager := mocks.NewAlertManager(ctrl)
	ethClient := mocks.NewMockEthClientInterface(ctrl)

	// NOTE - These tests should be migrated to the subsystem manager package
	// TODO(#76): No Subsystem Manager Tests

	ctx := context.Background()

	ctx = context.WithValue(ctx, core.L1Client, ethClient)
	ctx = context.WithValue(ctx, core.L2Client, ethClient)

	m := subsystem.NewManager(ctx, etlManager, engineManager, alertManager)

	service := svc.New(ctx, &cfg, m)
	return testSuite{
		testCfg: cfg,

		mockAlertMan:           alertManager,
		mockEngineMan:          engineManager,
		mockEtlMan:             etlManager,
		mockEthClientInterface: ethClient,

		apiSvc:   service,
		mockCtrl: ctrl,
	}
}
