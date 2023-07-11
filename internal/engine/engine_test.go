package engine_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testSuite struct {
	ctrl    *gomock.Controller
	re      engine.RiskEngine
	mockInv *mocks.MockInvariant
}

func createTestSuite(t *testing.T) *testSuite {
	ctrl := gomock.NewController(t)

	return &testSuite{
		ctrl:    ctrl,
		re:      engine.NewHardCodedEngine(),
		mockInv: mocks.NewMockInvariant(ctrl),
	}
}

func testErr() error {
	return fmt.Errorf("test error")
}

func Test_HardCodedEngine(t *testing.T) {
	var tests = []struct {
		name string
		test func(t *testing.T, ts *testSuite)
	}{
		{
			name: "Invalidation Failure From Error",
			test: func(t *testing.T, ts *testSuite) {
				td := core.TransitData{}

				ts.mockInv.EXPECT().Invalidate(td).
					Return(nil, false, testErr()).Times(1)

				ts.mockInv.EXPECT().SUUID().
					Return(core.NilSUUID()).Times(1)

				outcome, invalid := ts.re.Execute(context.Background(), td, ts.mockInv)
				assert.Nil(t, outcome)
				assert.False(t, invalid)

			}},
		{
			name: "Successful Invalidation",
			test: func(t *testing.T, ts *testSuite) {
				td := core.TransitData{}

				expectedOut := &core.InvalOutcome{
					Message: "20 inch blade on the Impala",
				}

				ts.mockInv.EXPECT().Invalidate(td).
					Return(expectedOut, true, nil).Times(1)

				ts.mockInv.EXPECT().SUUID().
					Return(core.NilSUUID()).Times(1)

				outcome, invalid := ts.re.Execute(context.Background(), td, ts.mockInv)
				assert.NotNil(t, outcome)
				assert.True(t, invalid)
				assert.Equal(t, expectedOut, outcome)
			}},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, test.name), func(t *testing.T) {
			ts := createTestSuite(t)
			test.test(t, ts)
		})
	}
}
