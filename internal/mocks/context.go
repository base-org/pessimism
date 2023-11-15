package mocks

import (
	"context"

	client "github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/state"
	gomock "github.com/golang/mock/gomock"
)

type MockSuite struct {
	Ctrl        *gomock.Controller
	Bundle      *client.Bundle
	MockIndexer *MockIxClient
	MockL1Node  *MockNodeClient
	MockL1      *MockEthClient
	MockL2      *MockEthClient
	MockL2Node  *MockNodeClient
	SS          state.Store
}

// Context ... Creates a context with mocked clients
func Context(ctx context.Context, ctrl *gomock.Controller) (context.Context, *MockSuite) {
	// 1. Construct mocked bundle
	mockedClient := NewMockEthClient(ctrl)
	mockedIndexer := NewMockIxClient(ctrl)
	mockedNode := NewMockNodeClient(ctrl)

	ss := state.NewMemState()

	bundle := &client.Bundle{
		IxClient: mockedIndexer,
		L1Client: mockedClient,
		L1Node:   mockedNode,
		L2Client: mockedClient,
		L2Node:   mockedNode,
	}

	// 2. Bind to context
	ctx = context.WithValue(ctx, core.State, ss)
	ctx = context.WithValue(ctx, core.Clients, bundle)

	// 3. Generate mock suite
	mockSuite := &MockSuite{
		Ctrl:        ctrl,
		Bundle:      bundle,
		MockIndexer: mockedIndexer,
		MockL1:      mockedClient,
		MockL1Node:  mockedNode,
		MockL2:      mockedClient,
		MockL2Node:  mockedNode,
		SS:          ss,
	}

	return ctx, mockSuite
}
