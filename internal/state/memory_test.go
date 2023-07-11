package state_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/state"
	"github.com/stretchr/testify/assert"
)

func Test_MemState(t *testing.T) {

	testKey := &core.StateKey{false, 1, "test", nil}
	testValue := "0xabc"

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
				_, err := ss.SetSlice(context.Background(), testKey, testValue)
				assert.NoError(t, err)

				val, err := ss.GetSlice(context.Background(), testKey)

				assert.NoError(t, err)
				assert.Equal(t, []string{testValue}, val)
			},
		},
		{
			name:         "Test_Get_Fail",
			description:  "Test failed get when key doens't exist",
			function:     "Get",
			construction: state.NewMemState,
			testLogic: func(t *testing.T, ss state.Store) {
				_, err := ss.GetSlice(context.Background(), testKey)
				assert.Error(t, err)
			},
		},
		{
			name:        "Test_Remove",
			description: "Test remove when value is prepopulated",
			function:    "Remove",
			construction: func() state.Store {
				ss := state.NewMemState()
				_, err := ss.SetSlice(context.Background(), testKey, testValue)
				if err != nil {
					panic(err)
				}

				return ss
			},
			testLogic: func(t *testing.T, ss state.Store) {
				err := ss.Remove(context.Background(), testKey)
				assert.NoError(t, err, "should not error")
			},
		},
	}

	// TODO - Consider making generic test helpers for this
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, test.name, test.function), func(t *testing.T) {
			testState := test.construction()
			test.testLogic(t, testState)
		})
	}
}
