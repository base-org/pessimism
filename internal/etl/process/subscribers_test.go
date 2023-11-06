package process

import (
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestAddRemoveSubscription(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		construction func() *subscribers
		test         func(*testing.T, *subscribers)
	}{
		{
			name:        "Successful Multi Add Test",
			description: "Many subscriptions should be addable",

			construction: func() *subscribers {
				return &subscribers{
					subs: make(map[core.ProcIdentifier]chan core.Event),
				}
			},

			test: func(t *testing.T, s *subscribers) {

				for _, id := range []core.ProcessID{
					core.MakeProcessID(1, 54, 43, 32),
					core.MakeProcessID(2, 54, 43, 32),
					core.MakeProcessID(3, 54, 43, 32),
					core.MakeProcessID(4, 54, 43, 32)} {
					outChan := make(chan core.Event)
					err := s.AddSubscriber(id, outChan)

					assert.NoError(t, err, "Ensuring that no error when adding new sub")

					_, exists := s.subs[id.ID]
					assert.True(t, exists, "Ensuring that key exists")
				}
			},
		},
		{
			name:        "Failed Add Test",
			description: "Duplicate subscribers cannot exist",

			construction: func() *subscribers {
				id := core.MakeProcessID(1, 54, 43, 32)
				outChan := make(chan core.Event)

				s := &subscribers{
					subs: make(map[core.ProcIdentifier]chan core.Event),
				}
				if err := s.AddSubscriber(id, outChan); err != nil {
					panic(err)
				}

				return s
			},

			test: func(t *testing.T, s *subscribers) {
				id := core.MakeProcessID(1, 54, 43, 32)
				outChan := make(chan core.Event)
				err := s.AddSubscriber(id, outChan)

				assert.Error(t, err, "Error was not generated when adding conflicting subs with same ID")
				assert.Equal(t, err.Error(), fmt.Sprintf(subExistsErr, id.String()), "Ensuring that returned error is a not found type")
			},
		},
		{
			name:        "Successful Remove Test",
			description: "Subscribers should be removable",

			construction: func() *subscribers {
				id := core.MakeProcessID(1, 54, 43, 32)
				outChan := make(chan core.Event)

				s := &subscribers{
					subs: make(map[core.ProcIdentifier]chan core.Event),
				}
				if err := s.AddSubscriber(id, outChan); err != nil {
					panic(err)
				}

				return s
			},

			test: func(t *testing.T, s *subscribers) {
				id := core.MakeProcessID(1, 54, 43, 32)
				err := s.RemoveSubscriber(id)

				assert.NoError(t, err, "Ensuring that no error is thrown when removing an existing sub")

				_, exists := s.subs[core.MakeProcessID(1, 54, 43, 32).ID]
				assert.False(t, exists, "Ensuring that key is removed from mapping")
			},
		}, {
			name:        "Failed Remove Test",
			description: "Unknown keys should not be removable",

			construction: func() *subscribers {
				id := core.MakeProcessID(1, 54, 43, 32)
				outChan := make(chan core.Event)

				s := &subscribers{
					subs: make(map[core.ProcIdentifier]chan core.Event),
				}
				if err := s.AddSubscriber(id, outChan); err != nil {
					panic(err)
				}

				return s
			},

			test: func(t *testing.T, s *subscribers) {

				id := core.MakeProcessID(69, 69, 69, 69)
				err := s.RemoveSubscriber(id)

				assert.Error(t, err, "Ensuring that an error is thrown when trying to remove a non-existent sub")
				assert.Equal(t, err.Error(), fmt.Sprintf(subNotFound, id.Identifier()))
			},
		},
		{
			name:        "Passed Engine Test",
			description: "When a relay is passed to AddRelay, it should be used during transit operations",

			construction: func() *subscribers {
				return &subscribers{}
			},
			test: func(t *testing.T, s *subscribers) {
				relayChan := make(chan core.HeuristicInput)

				PathID := core.PathID{}

				relay := core.NewEngineRelay(PathID, relayChan)

				err := s.AddEngineRelay(relay)
				assert.NoError(t, err)

				testData := core.Event{Network: 2, Value: "goodbye closed-source blocksec monitoring"}
				expectedInput := core.HeuristicInput{
					PathID: PathID,
					Input:  testData,
				}

				go func(t *testing.T) {
					assert.NoError(t, s.Publish(testData))
				}(t)

				actualInput := <-relayChan

				assert.Equal(t, actualInput, expectedInput)

			},
		},
		{
			name:        "Failed Engine Test",
			description: "When relay already exists and AddRelay function is called, an error should be returned",

			construction: func() *subscribers {
				relayChan := make(chan core.HeuristicInput)

				relay := core.NewEngineRelay(core.PathID{}, relayChan)
				s := &subscribers{
					subs: make(map[core.ProcIdentifier]chan core.Event),
				}

				if err := s.AddEngineRelay(relay); err != nil {
					panic(err)
				}

				return s
			},

			test: func(t *testing.T, s *subscribers) {
				relayChan := make(chan core.HeuristicInput)

				relay := core.NewEngineRelay(core.PathID{}, relayChan)

				err := s.AddEngineRelay(relay)

				assert.Error(t, err)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			subs := tc.construction()
			tc.test(t, subs)
		})

	}
}

func TestPublishToSubscribers(t *testing.T) {
	s := &subscribers{
		subs: make(map[core.ProcIdentifier]chan core.Event),
	}

	var subs = []struct {
		channel chan core.Event
		id      core.ProcessID
	}{
		{
			channel: make(chan core.Event, 1),
			id:      core.MakeProcessID(3, 54, 43, 32),
		},
		{
			channel: make(chan core.Event, 1),
			id:      core.MakeProcessID(1, 54, 43, 32),
		},
		{
			channel: make(chan core.Event, 1),
			id:      core.MakeProcessID(1, 2, 43, 32),
		},
		{
			channel: make(chan core.Event, 1),
			id:      core.MakeProcessID(1, 4, 43, 32),
		},
	}

	for _, sub := range subs {
		err := s.AddSubscriber(sub.id, sub.channel)
		assert.NoError(t, err, "Received error when trying to add sub")
	}

	expected := core.Event{
		Timestamp: time.Date(1969, time.April, 1, 4, 20, 0, 0, time.Local),
		Type:      3,
		Value:     0x42069,
	}

	err := s.Publish(expected)
	assert.NoError(t, err, "Received error when trying to transit output")

	for _, sub := range subs {
		actual := <-sub.channel

		assert.Equal(t, actual, expected, "Ensuring transited data is actually returned on channels used by sub")
	}

}
