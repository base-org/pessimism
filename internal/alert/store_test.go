package alert_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/core"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	var tests = []struct {
		name        string
		description string
		testLogic   func(t *testing.T)
	}{
		{
			name:        "Test Get Alert Policy Success",
			description: "Test GetAlertPolicy",
			testLogic: func(t *testing.T) {
				am := alert.NewStore()

				id := core.UUID{}
				policy := &core.AlertPolicy{
					Msg:  "test message",
					Dest: core.Slack.String(),
				}

				err := am.AddAlertPolicy(id, policy)
				assert.NoError(t, err)

				actualPolicy, err := am.GetAlertPolicy(id)
				assert.NoError(t, err)
				assert.Equal(t, policy, actualPolicy)
			},
		},
		{
			name:        "Test Add Alert Policy Success",
			description: "Test adding of arbitrary Alert Policies",
			testLogic: func(t *testing.T) {
				am := alert.NewStore()

				id := core.UUID{}
				policy := &core.AlertPolicy{
					Dest: core.Slack.String(),
				}

				err := am.AddAlertPolicy(id, policy)
				assert.NoError(t, err)

				// add again
				err = am.AddAlertPolicy(id, policy)
				assert.Error(t, err)
			},
		},
		{
			name:        "Test NewStore",
			description: "Test NewStore logic",
			testLogic: func(t *testing.T) {
				am := alert.NewStore()
				assert.NotNil(t, am)
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%s:%d", test.name, i), func(t *testing.T) {
			test.testLogic(t)
		})
	}
}
