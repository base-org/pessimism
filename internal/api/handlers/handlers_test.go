package handlers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
)

func testSUUID1() core.SUUID {
	return core.MakeSUUID(1, 1, 1)
}

func testError1() error {
	return fmt.Errorf("test error 1")
}

type testSuite struct {
	mockSvc mocks.MockService

	testHandler handlers.Handlers
	ctrl        *gomock.Controller
}

func createTestSuite(t *testing.T) testSuite {
	ctrl := gomock.NewController(t)

	mockSvc := mocks.NewMockService(ctrl)
	testHandler, err := handlers.New(context.Background(), mockSvc)

	if err != nil {
		panic(err)
	}

	return testSuite{
		mockSvc:     *mockSvc,
		testHandler: testHandler,
		ctrl:        ctrl,
	}
}
