package component

import (
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_Add_Remove_Directive(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		constructionLogic func() *router
		testLogic         func(*testing.T, *router)
	}{
		{
			name:        "Successful Multi Add Test",
			description: "When multiple directives are passed to AddDirective function, they should successfully be added to the router mapping",

			constructionLogic: func() *router {
				router, _ := newRouter()
				return router
			},

			testLogic: func(t *testing.T, router *router) {

				for _, id := range []core.ComponentID{
					core.MakeComponentID(1, 54, 43, 32),
					core.MakeComponentID(2, 54, 43, 32),
					core.MakeComponentID(3, 54, 43, 32),
					core.MakeComponentID(4, 54, 43, 32)} {
					outChan := make(chan core.TransitData)
					err := router.AddDirective(id, outChan)

					assert.NoError(t, err, "Ensuring that no error when adding new directive")

					_, exists := router.outChans[id]
					assert.True(t, exists, "Ensuring that key exists")
				}
			},
		},
		{
			name:        "Failed Add Test",
			description: "When existing directive is passed to AddDirective function it should fail to be added to the router mapping",

			constructionLogic: func() *router {
				id := core.MakeComponentID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)

				router, _ := newRouter()
				_ = router.AddDirective(id, outChan)
				return router
			},

			testLogic: func(t *testing.T, router *router) {
				id := core.MakeComponentID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)
				err := router.AddDirective(id, outChan)

				assert.Error(t, err, "Error was not generated when adding conflicting directives with same ID")
				assert.Equal(t, err.Error(), fmt.Sprintf(dirAlreadyExistsErr, id.String()), "Ensuring that returned error is a not found type")
			},
		},
		{
			name:        "Successful Remove Test",
			description: "When existing directive is passed to RemoveDirective function, it should be removed from mapping",

			constructionLogic: func() *router {
				id := core.MakeComponentID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)

				router, _ := newRouter()
				_ = router.AddDirective(id, outChan)
				return router
			},

			testLogic: func(t *testing.T, router *router) {

				err := router.RemoveDirective(core.MakeComponentID(1, 54, 43, 32))

				assert.NoError(t, err, "Ensuring that no error is thrown when removing an existing directive")

				_, exists := router.outChans[core.MakeComponentID(1, 54, 43, 32)]
				assert.False(t, exists, "Ensuring that key is removed from mapping")
			},
		}, {
			name:        "Failed Remove Test",
			description: "When non-existing directive key is passed to RemoveDirective function, an error should be returned",

			constructionLogic: func() *router {
				id := core.MakeComponentID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)

				router, _ := newRouter()
				_ = router.AddDirective(id, outChan)
				return router
			},

			testLogic: func(t *testing.T, router *router) {

				err := router.RemoveDirective(core.MakeComponentID(69, 69, 69, 69))

				assert.Error(t, err, "Ensuring that an error is thrown when trying to remove a non-existent directive")
				assert.Equal(t, err.Error(), fmt.Sprintf(dirNotFoundErr, core.MakeComponentID(69, 69, 69, 69).String()))
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			testRouter := tc.constructionLogic()
			tc.testLogic(t, testRouter)
		})

	}
}

func Test_Transit_Output(t *testing.T) {
	testRouter, _ := newRouter()

	var directives = []struct {
		channel chan core.TransitData
		id      core.ComponentID
	}{
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeComponentID(3, 54, 43, 32),
		},
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeComponentID(1, 54, 43, 32),
		},
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeComponentID(1, 2, 43, 32),
		},
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeComponentID(1, 4, 43, 32),
		},
	}

	for _, directive := range directives {
		err := testRouter.AddDirective(directive.id, directive.channel)
		assert.NoError(t, err, "Received error when trying to add directive")
	}

	expectedOutput := core.TransitData{
		Timestamp: time.Date(1969, time.April, 1, 4, 20, 0, 0, time.Local),
		Type:      3,
		Value:     0x42069,
	}

	err := testRouter.TransitOutput(expectedOutput)
	assert.NoError(t, err, "Receieved error when trying to transit output")

	for _, directive := range directives {
		actualOutput := <-directive.channel

		assert.Equal(t, actualOutput, expectedOutput, "Ensuring transited data is actually returned on channels used by Router")
	}

}
