package component

import (
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_Add_Remove_Egress(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		constructionLogic func() *egressHandler
		testLogic         func(*testing.T, *egressHandler)
	}{
		{
			name:        "Successful Multi Add Test",
			description: "When multiple egresses are passed to AddEgress function, they should successfully be added to the egress mapping",

			constructionLogic: func() *egressHandler {
				handler := newEgressHandler()
				return handler
			},

			testLogic: func(t *testing.T, eh *egressHandler) {

				for _, id := range []core.CUUID{
					core.MakeCUUID(1, 54, 43, 32),
					core.MakeCUUID(2, 54, 43, 32),
					core.MakeCUUID(3, 54, 43, 32),
					core.MakeCUUID(4, 54, 43, 32)} {
					outChan := make(chan core.TransitData)
					err := eh.AddEgress(id, outChan)

					assert.NoError(t, err, "Ensuring that no error when adding new egress")

					_, exists := eh.egresses[id.PID]
					assert.True(t, exists, "Ensuring that key exists")
				}
			},
		},
		{
			name:        "Failed Add Test",
			description: "When existing direcegresstive is passed to AddEgress function it should fail to be added to the egress mapping",

			constructionLogic: func() *egressHandler {
				id := core.MakeCUUID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)

				handler := newEgressHandler()
				if err := handler.AddEgress(id, outChan); err != nil {
					panic(err)
				}

				return handler
			},

			testLogic: func(t *testing.T, eh *egressHandler) {
				id := core.MakeCUUID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)
				err := eh.AddEgress(id, outChan)

				assert.Error(t, err, "Error was not generated when adding conflicting egresses with same ID")
				assert.Equal(t, err.Error(), fmt.Sprintf(egressAlreadyExistsErr, id.String()), "Ensuring that returned error is a not found type")
			},
		},
		{
			name:        "Successful Remove Test",
			description: "When existing egress is passed to RemoveEgress function, it should be removed from mapping",

			constructionLogic: func() *egressHandler {
				id := core.MakeCUUID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)

				handler := newEgressHandler()
				if err := handler.AddEgress(id, outChan); err != nil {
					panic(err)
				}

				return handler
			},

			testLogic: func(t *testing.T, eh *egressHandler) {

				err := eh.RemoveEgress(core.MakeCUUID(1, 54, 43, 32))

				assert.NoError(t, err, "Ensuring that no error is thrown when removing an existing egress")

				_, exists := eh.egresses[core.MakeCUUID(1, 54, 43, 32).PID]
				assert.False(t, exists, "Ensuring that key is removed from mapping")
			},
		}, {
			name:        "Failed Remove Test",
			description: "When non-existing egress key is passed to RemoveEgress function, an error should be returned",

			constructionLogic: func() *egressHandler {
				id := core.MakeCUUID(1, 54, 43, 32)
				outChan := make(chan core.TransitData)

				handler := newEgressHandler()
				if err := handler.AddEgress(id, outChan); err != nil {
					panic(err)
				}

				return handler
			},

			testLogic: func(t *testing.T, eh *egressHandler) {

				cID := core.MakeCUUID(69, 69, 69, 69)
				err := eh.RemoveEgress(cID)

				assert.Error(t, err, "Ensuring that an error is thrown when trying to remove a non-existent egress")
				assert.Equal(t, err.Error(), fmt.Sprintf(egressNotFoundErr, cID.PID.String()))
			},
		},
		{
			name:        "Passed Engine Egress Test",
			description: "When a relay is passed to AddRelay, it should be used during transit operations",

			constructionLogic: newEgressHandler,

			testLogic: func(t *testing.T, eh *egressHandler) {
				relayChan := make(chan core.InvariantInput)

				pUUID := core.NilPUUID()

				relay := core.NewEngineRelay(pUUID, relayChan)

				handler := newEgressHandler()

				err := handler.AddRelay(relay)
				assert.NoError(t, err)

				testData := core.TransitData{Network: 2, Value: "goodbye closed-source blocksec monitoring"}
				expectedInput := core.InvariantInput{
					PUUID: pUUID,
					Input: testData,
				}

				go func(t *testing.T) {
					assert.NoError(t, handler.Send(testData))
				}(t)

				actualInput := <-relayChan

				assert.Equal(t, actualInput, expectedInput)

			},
		},
		{
			name:        "Failed Engine Egress Test",
			description: "When relay already exists and AddRelay function is called, an error should be returned",

			constructionLogic: func() *egressHandler {
				relayChan := make(chan core.InvariantInput)

				pUUID := core.NilPUUID()

				relay := core.NewEngineRelay(pUUID, relayChan)

				handler := newEgressHandler()

				if err := handler.AddRelay(relay); err != nil {
					panic(err)
				}

				return handler
			},

			testLogic: func(t *testing.T, eh *egressHandler) {
				relayChan := make(chan core.InvariantInput)

				pUUID := core.NilPUUID()

				relay := core.NewEngineRelay(pUUID, relayChan)

				err := eh.AddRelay(relay)

				assert.Error(t, err)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			testegress := tc.constructionLogic()
			tc.testLogic(t, testegress)
		})

	}
}

func Test_Transit_Output(t *testing.T) {
	testHandler := newEgressHandler()

	var egresses = []struct {
		channel chan core.TransitData
		id      core.CUUID
	}{
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeCUUID(3, 54, 43, 32),
		},
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeCUUID(1, 54, 43, 32),
		},
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeCUUID(1, 2, 43, 32),
		},
		{
			channel: make(chan core.TransitData, 1),
			id:      core.MakeCUUID(1, 4, 43, 32),
		},
	}

	for _, egress := range egresses {
		err := testHandler.AddEgress(egress.id, egress.channel)
		assert.NoError(t, err, "Received error when trying to add egress")
	}

	expectedOutput := core.TransitData{
		Timestamp: time.Date(1969, time.April, 1, 4, 20, 0, 0, time.Local),
		Type:      3,
		Value:     0x42069,
	}

	err := testHandler.Send(expectedOutput)
	assert.NoError(t, err, "Receieved error when trying to transit output")

	for _, egress := range egresses {
		actualOutput := <-egress.channel

		assert.Equal(t, actualOutput, expectedOutput, "Ensuring transited data is actually returned on channels used by egress")
	}

}
