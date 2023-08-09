package pipe_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry/pipe"
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

	p, err := pipe.NewEventDefinition(ctx, core.Layer1)

	if err != nil {
		t.Fatal(err)
	}

	return &testSuite{
		ctx:       ctx,
		def:       p,
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

				_, err := suite.def.Transform(suite.ctx, core.TransitData{})
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
