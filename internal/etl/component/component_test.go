package component

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_MetaData(t *testing.T) {
	var tests = []struct {
		name        string
		description string
		function    string

		constructionLogic func() *metaData
		testLogic         func(*testing.T, *metaData)
	}{
		{
			name:        "Test State Change Emit",
			description: "When emitStateChange is called, a new state should be state for metadata and sent in channel",
			function:    "emitStateChange",

			constructionLogic: func() *metaData {
				return newMetaData(0, 0)
			},

			testLogic: func(t *testing.T, md *metaData) {

				go func() {
					// Simulate a component ending itself
					md.emitStateChange(Terminated)

				}()

				sChange := <-md.stateChan

				assert.Equal(t, sChange.From, Inactive)
				assert.Equal(t, sChange.To, Terminated)
				assert.Equal(t, sChange.ID, core.NilCUUID())
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
