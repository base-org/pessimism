package mocks

import (
	context "context"

	client "github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/state"
	gomock "github.com/golang/mock/gomock"
)

type MockSuite struct {
	Ctrl   *gomock.Controller
	Bundle *client.Bundle
	MockL1 *MockEthClient
	MockL2 *MockEthClient
	SS     state.Store
}

// Context ... Creates a context with mocked clients
func Context(ctx context.Context, ctrl *gomock.Controller) (context.Context, *MockSuite) {
	// 1. Construct mocked bundle
	mockedClient := NewMockEthClient(ctrl)
	ss := state.NewMemState()

	bundle := &client.Bundle{
		L1Client: mockedClient,
		L2Client: mockedClient,
	}

	// 2. Bind to context
	ctx = context.WithValue(ctx, core.State, ss)
	ctx = context.WithValue(ctx, core.Clients, bundle)

	// 3. Generate mock suite
	mockSuite := &MockSuite{
		Ctrl:   ctrl,
		Bundle: bundle,
		MockL1: mockedClient,
		MockL2: mockedClient,
		SS:     ss,
	}

	return ctx, mockSuite
}
