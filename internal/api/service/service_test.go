package service_test

import (
	"context"
	"fmt"

	svc "github.com/base-org/pessimism/internal/api/service"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/client"

	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
)

const (
	testErrMsg1 = "69"
	testErrMsg2 = "420"
	testErrMsg3 = "666"
)

type testSuite struct {
	apiSvc     svc.Service
	mockClient *mocks.MockEthClient
	mockSub    *mocks.SubManager
	mockCtrl   *gomock.Controller
}

func testErr1() error {
	return fmt.Errorf(testErrMsg1)
}

func createTestSuite(ctrl *gomock.Controller) *testSuite {
	sysMock := mocks.NewSubManager(ctrl)
	ethClient := mocks.NewMockEthClient(ctrl)
	ctx := context.Background()

	ctx = app.InitializeContext(ctx, nil, &client.Bundle{
		L1Client: ethClient,
		L2Client: ethClient,
	})

	service := svc.New(ctx, sysMock)
	return &testSuite{
		apiSvc:     service,
		mockClient: ethClient,
		mockSub:    sysMock,
		mockCtrl:   ctrl,
	}
}
