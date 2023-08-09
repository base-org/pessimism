package pipe_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry/pipe"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/mocks"
)

// testSuite ... Test suite for the event log pipe
type testSuite struct {
	ctx       context.Context
	def       component.PipeDefinition
	mockSuite *mocks.MockSuite
}

// defConstructor ... Default constructor for the test suite
func defConstructor(t *testing.T) *testSuite {
	ctrl := gomock.NewController(t)
	ctx, suite := mocks.Context(context.Background(), ctrl)

	ed, err := pipe.NewEventDefinition(ctx, core.Layer1)
	if err != nil {
		t.Fatal(err)
	}

	nilKey := &core.StateKey{}
	ed.SK = nilKey

	return &testSuite{
		ctx:       ctx,
		def:       ed,
		mockSuite: suite,
	}
}

// TestEventLogPipe ... Tests the event log pipe
func TestEventLogPipe(t *testing.T) {
	var tests = []struct {
		name        string
		constructor func(t *testing.T) *testSuite
		runner      func(t *testing.T, suite *testSuite)
	}{
		{
			name:        "No Error When no Events to Monitor",
			constructor: defConstructor,
			runner: func(t *testing.T, suite *testSuite) {
				_, err := suite.def.Transform(suite.ctx, core.TransitData{
					Value: types.Block{},
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "No Error When no Events to Monitor",
			constructor: func(t *testing.T) *testSuite {
				ts := defConstructor(t)

				state.InsertUnique(ts.ctx, &core.StateKey{
					Nesting: true,
				}, "0x00000000")

				innerKey := &core.StateKey{
					Nesting: false,
					ID:      "0x00000000",
				}

				state.InsertUnique(ts.ctx, innerKey, "transfer(address,address,uint256)")
				return ts
			},
			runner: func(t *testing.T, suite *testSuite) {
				suite.mockSuite.MockL1.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return(nil, nil)

				_, err := suite.def.Transform(suite.ctx, core.TransitData{
					Value: types.NewBlockWithHeader(&types.Header{}),
				})
				assert.NoError(t, err)
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
