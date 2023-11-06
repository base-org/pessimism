package registry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/mocks"
)

// testSuite ... Test suite for the event log subscription
type testSuite struct {
	ctx       context.Context
	def       process.Subscription
	mockSuite *mocks.MockSuite
}

// defConstructor ... Default constructor for the test suite
func defConstructor(t *testing.T) *testSuite {
	ctrl := gomock.NewController(t)
	ctx, suite := mocks.Context(context.Background(), ctrl)

	// Populate the state store with the events to monitor
	// NOTE - There's likely a more extensible way to handle nested keys in the state store
	_ = state.InsertUnique(ctx, &core.StateKey{
		Nesting: true,
	}, "0x00000000")

	innerKey := &core.StateKey{
		Nesting: false,
		ID:      "0x00000000",
	}

	_ = state.InsertUnique(ctx, innerKey, "transfer(address,address,uint256)")

	subscript, err := registry.NewLogSubscript(ctx, core.Layer1)
	if err != nil {
		t.Fatal(err)
	}

	subscript.SK = &core.StateKey{}

	return &testSuite{
		ctx:       ctx,
		def:       subscript,
		mockSuite: suite,
	}
}

// TestLogSubscription ... Tests the event log subscription
func TestLogSubscription(t *testing.T) {
	var tests = []struct {
		name        string
		constructor func(t *testing.T) *testSuite
		runner      func(t *testing.T, suite *testSuite)
	}{
		{
			name:        "Error when failed FilterQuery",
			constructor: defConstructor,
			runner: func(t *testing.T, ts *testSuite) {
				ts.mockSuite.MockL1.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("unknown block")).
					Times(10)

				_, err := ts.def.Run(ts.ctx, core.Event{
					Value: types.Header{}})
				assert.Error(t, err)
			},
		},
		{
			name: "No Error When Successful Filter Query",
			constructor: func(t *testing.T) *testSuite {
				ts := defConstructor(t)

				return ts
			},
			runner: func(t *testing.T, ts *testSuite) {
				ts.mockSuite.MockL1.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return(nil, nil)

				tds, err := ts.def.Run(ts.ctx, core.Event{
					Value: types.Header{},
				})
				assert.NoError(t, err)
				assert.Empty(t, tds)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := tt.constructor(t)
			tt.runner(t, suite)
		})
	}

}
