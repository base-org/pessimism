package pipeline

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
)

func Test_AddDirective(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		constructionLogic func() *OutputRouter
		testLogic         func(*testing.T, *OutputRouter)
	}{

		{
			name:        "Successful Add Test",
			description: "When new directive is passed it should be added to router mapping",

			constructionLogic: func() *OutputRouter {
				return NewOutputRouter()
			},

			testLogic: func(t *testing.T, router *OutputRouter) {
				id := xid.New()
				outChan := make(chan models.TransitData)
				err := router.AddDirective(id, outChan)

				assert.NoError(t, err, "Ensuring that no error when adding new directive")

				_, exists := router.outChans[id]
				assert.True(t, exists, "Ensuring that key exists")
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			router := tc.constructionLogic()
			tc.testLogic(t, router)
		})

	}

	// TODO - Test other add directive cases
}

// TODO - Test Transit Output function
