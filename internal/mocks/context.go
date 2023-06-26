package mocks

import (
	context "context"

	"github.com/base-org/pessimism/internal/core"
	gomock "github.com/golang/mock/gomock"
)

// Context ... Creates a context with mocked clients
func Context(ctx context.Context, ctrl *gomock.Controller) context.Context {
	mockedL1Client := NewMockEthClientInterface(ctrl)
	mockedL2Client := NewMockEthClientInterface(ctrl)

	ctx = context.WithValue(ctx, core.L1Client, mockedL1Client)
	ctx = context.WithValue(ctx, core.L2Client, mockedL2Client)

	return ctx
}
