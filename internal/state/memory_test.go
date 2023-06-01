package state_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/state"
	"github.com/stretchr/testify/assert"
)

func Test_MemState(t *testing.T) {

	var tests = []struct {
		name        string
		description string
		function    string

		construction func() state.Store
		testLogic    func(t *testing.T, ss state.Store)
	}{
		{
			name:         "Test_Set_Success",
			description:  "Test set",
			function:     "Set",
			construction: state.NewMemState,
			testLogic: func(t *testing.T, ss state.Store) {
				_, err := ss.Set(context.Background(), "test", "test")
				assert.NoError(t, err)

				val, err := ss.Get(context.Background(), "test")

				assert.NoError(t, err)
				assert.Equal(t, []string{"test"}, val)
			},
		},
		{
			name:         "Test_Get_Fail",
			description:  "Test failed get when key doens't exist",
			function:     "Get",
			construction: state.NewMemState,
			testLogic: func(t *testing.T, ss state.Store) {
				_, err := ss.Get(context.Background(), "test")
				assert.Error(t, err)
			},
		},
		{
			name:        "Test_Get_Success",
			description: "Test set when value is prepopulated",
			function:    "Get",
			construction: func() state.Store {
				ss := state.NewMemState()
				_, err := ss.Set(context.Background(), "0x123", "0xabc")
				if err != nil {
					panic(err)
				}

				return ss
			},
			testLogic: func(t *testing.T, ss state.Store) {
				_, err := ss.Get(context.Background(), "test")
				assert.Error(t, err)
			},
		},
		{
			name:        "Test_Remove",
			description: "Test remove when value is prepopulated",
			function:    "Remove",
			construction: func() state.Store {
				ss := state.NewMemState()
				_, err := ss.Set(context.Background(), "0x123", "0xabc")
				if err != nil {
					panic(err)
				}

				return ss
			},
			testLogic: func(t *testing.T, ss state.Store) {
				err := ss.Remove(context.Background(), "0x123")
				assert.NoError(t, err, "should not error")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testState := test.construction()
			test.testLogic(t, testState)
		})
	}
}
