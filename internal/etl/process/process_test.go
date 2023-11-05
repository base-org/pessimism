package process

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_State(t *testing.T) {
	var tests = []struct {
		name        string
		description string
		function    string

		constructionLogic func() *State
		testLogic         func(*testing.T, *State)
	}{
		{
			name:        "Test State Change Emit",
			description: "When emitStateChange is called, a new state should be state for State and sent in channel",
			function:    "emitStateChange",

			constructionLogic: func() *State {
				return newState(0, 0)
			},

			testLogic: func(t *testing.T, s *State) {

				go func() {
					// Simulate a process ending itself
					s.emit(Terminated)

				}()

				s1 := <-s.relay

				assert.Equal(t, s1.From, Inactive)
				assert.Equal(t, s1.To, Terminated)
				assert.Equal(t, s1.ID, core.ProcessID{})
			},
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			testMeta := tc.constructionLogic()
			tc.testLogic(t, testMeta)
		})

	}
}
