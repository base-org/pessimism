package registry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/etl/registry"
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

	ed, err := registry.NewEventDefinition(ctx, core.Layer1)
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
			name:        "Error when failed FilterQuery",
			constructor: defConstructor,
			runner: func(t *testing.T, suite *testSuite) {
				suite.mockSuite.MockL1.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("unknown block"))

				_, err := suite.def.Transform(suite.ctx, core.TransitData{
					Value: *types.NewBlockWithHeader(&types.Header{})})
				assert.Error(t, err)
			},
		},
		{
			name: "No Error When Successful Filter Query",
			constructor: func(t *testing.T) *testSuite {
				ts := defConstructor(t)

				return ts
			},
			runner: func(t *testing.T, suite *testSuite) {
				suite.mockSuite.MockL1.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return(nil, nil)

				tds, err := suite.def.Transform(suite.ctx, core.TransitData{
					Value: *types.NewBlockWithHeader(&types.Header{}),
				})
				assert.NoError(t, err)
				assert.Empty(t, tds)
			},
		},
		{
			name: "DLQ Retry When Failed Filter Query",
			constructor: func(t *testing.T) *testSuite {
				ts := defConstructor(t)

				return ts
			},
			runner: func(t *testing.T, suite *testSuite) {
				// 1. Fail the first filter query and assert that the DLQ is populated
				suite.mockSuite.MockL1.EXPECT().
					FilterLogs(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("unknown block"))

				tds, err := suite.def.Transform(suite.ctx, core.TransitData{
					Value: *types.NewBlockWithHeader(&types.Header{}),
				})
				assert.Error(t, err)
				assert.Empty(t, tds)

				log1 := types.Log{
					Address: common.HexToAddress("0x0"),
				}

				log2 := types.Log{
					Address: common.HexToAddress("0x1"),
				}

				// 2. Retry the filter query and assert that the DLQ is empty
				suite.mockSuite.MockL1.EXPECT().
					FilterLogs(gomock.Any(), gomock.Any()).
					Return([]types.Log{log1}, nil)

				suite.mockSuite.MockL1.EXPECT().
					FilterLogs(gomock.Any(), gomock.Any()).
					Return([]types.Log{log2}, nil)

				tds, err = suite.def.Transform(suite.ctx, core.TransitData{
					Value: *types.NewBlockWithHeader(&types.Header{}),
				})

				assert.NoError(t, err)
				assert.NotEmpty(t, tds)

				actualLog1, ok := tds[0].Value.(types.Log)
				assert.True(t, ok)

				actualLog2, ok := tds[1].Value.(types.Log)
				assert.True(t, ok)

				assert.Equal(t, actualLog1, log1)
				assert.Equal(t, actualLog2, log2)
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
