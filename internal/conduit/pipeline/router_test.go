package pipeline

import (
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/stretchr/testify/assert"
)

func Test_Add_Remove_Directive(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		constructionLogic func() *OutputRouter
		testLogic         func(*testing.T, *OutputRouter)
	}{

		{
			name:        "Successful Multi Add Test",
			description: "When multiple directives are passed to AddDirective function, they should successfully be added to the router mapping",

			constructionLogic: func() *OutputRouter {
				router, _ := NewOutputRouter()
				return router
			},

			testLogic: func(t *testing.T, router *OutputRouter) {

				for _, id := range []int{0x420, 0x42, 0x69, 0x666} {
					outChan := make(chan models.TransitData)
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

			constructionLogic: func() *OutputRouter {
				id := 0x420
				outChan := make(chan models.TransitData)

				router, _ := NewOutputRouter()
				_ = router.AddDirective(id, outChan)
				return router
			},

			testLogic: func(t *testing.T, router *OutputRouter) {
				id := 0x420
				outChan := make(chan models.TransitData)
				err := router.AddDirective(id, outChan)

				assert.Error(t, err, "Ensuring that no error is generated when adding new directive")
				assert.Equal(t, err.Error(), fmt.Sprintf(alreadyExistsErr, 0x420), "Ensuring that returned error is a not found type")
			},
		},
		{
			name:        "Successful Remove Test",
			description: "When existing directive is passed to RemoveDirective function, it should be removed from mapping",

			constructionLogic: func() *OutputRouter {
				id := 0x420
				outChan := make(chan models.TransitData)

				router, _ := NewOutputRouter()
				_ = router.AddDirective(id, outChan)
				return router
			},

			testLogic: func(t *testing.T, router *OutputRouter) {

				err := router.RemoveDirective(0x420)

				assert.NoError(t, err, "Ensuring that no error is thrown when removing an existing directive")

				_, exists := router.outChans[0x420]
				assert.False(t, exists, "Ensuring that key is removed from mapping")
			},
		}, {
			name:        "Failed Remove Test",
			description: "When non-existing directive key is passed to RemoveDirective function, an error should be returned",

			constructionLogic: func() *OutputRouter {
				id := 0x420
				outChan := make(chan models.TransitData)

				router, _ := NewOutputRouter()
				_ = router.AddDirective(id, outChan)
				return router
			},

			testLogic: func(t *testing.T, router *OutputRouter) {

				err := router.RemoveDirective(0x69)

				assert.Error(t, err, "Ensuring that an error is thrown when trying to remove a non-existent directive")
				assert.Equal(t, err.Error(), fmt.Sprintf(notFoundErr, 0x69))
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
	testRouter, _ := NewOutputRouter()

	var directives = []struct {
		channel chan models.TransitData
		id      int
	}{
		{
			channel: make(chan models.TransitData, 1),
			id:      0x420,
		},
		{
			channel: make(chan models.TransitData, 1),
			id:      0x42,
		},
		{
			channel: make(chan models.TransitData, 1),
			id:      0x69,
		},
		{
			channel: make(chan models.TransitData, 1),
			id:      0x666,
		},
	}

	for _, directive := range directives {
		err := testRouter.AddDirective(directive.id, directive.channel)
		assert.NoError(t, err, "Received error when trying to add directive")
	}

	expectedOutput := models.TransitData{
		Timestamp: time.Date(1969, time.April, 1, 4, 20, 0, 0, time.Local),
		Type:      "String Beanz",
		Value:     0x42069,
	}

	testRouter.TransitOutput(expectedOutput)

	for _, directive := range directives {
		actualOutput := <-directive.channel

		assert.Equal(t, actualOutput, expectedOutput, "Ensuring transited data is actually returned on channels used by OutputRouter")
	}

}
