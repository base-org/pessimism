package mocks

import (
	context "context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/state"
	gomock "github.com/golang/mock/gomock"
)

type MockSuite struct {
	Ctrl   *gomock.Controller
	MockL1 *MockEthClient
	MockL2 *MockEthClient
	SS     state.Store
}

// Context ... Creates a context with mocked clients
func Context(ctx context.Context, ctrl *gomock.Controller) (context.Context, *MockSuite) {
	// 1. Construct mocked clients
	mockedClient := NewMockEthClient(ctrl)
	ss := state.NewMemState()

	// 2. Bind to context
	ctx = context.WithValue(ctx, core.L1Client, mockedClient)
	ctx = context.WithValue(ctx, core.L2Client, mockedClient)
	ctx = context.WithValue(ctx, core.State, ss)

	// 3. Generate mock suite
	mockSuite := &MockSuite{
		Ctrl:   ctrl,
		MockL1: mockedClient,
		MockL2: mockedClient,
		SS:     ss,
	}

	return ctx, mockSuite
}
