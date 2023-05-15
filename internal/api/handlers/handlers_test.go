package handlers_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
)

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
