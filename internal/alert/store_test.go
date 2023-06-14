package alert_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_Store(t *testing.T) {
	var tests = []struct {
		name        string
		description string
		testLogic   func(t *testing.T)
	}{
		{
			name:        "Test Get Alert Destintation Success",
			description: "Test GetAlertDestination",
			testLogic: func(t *testing.T) {
				am := alert.NewStore()

				sUUID := core.MakeSUUID(core.Layer1, core.Live, core.BalanceEnforcement)
				alertDestination := core.Slack

				err := am.AddAlertDestination(sUUID, alertDestination)
				assert.NoError(t, err, "failed to add alert destination")

				actualAlertDest, err := am.GetAlertDestination(sUUID)
				assert.NoError(t, err, "failed to get alert destination")
				assert.Equal(t, alertDestination, actualAlertDest, "alert destination mismatch")
			},
		},
		{
			name:        "Test Add Alert Destination Success",
			description: "Test adding of arbitrary alert destinations",
			testLogic: func(t *testing.T) {
				am := alert.NewStore()

				sUUID := core.MakeSUUID(core.Layer1, core.Live, core.BalanceEnforcement)
				alertDestination := core.Slack

				err := am.AddAlertDestination(sUUID, alertDestination)
				assert.NoError(t, err, "failed to add alert destination")

				// add again
				err = am.AddAlertDestination(sUUID, alertDestination)
				assert.Error(t, err, "failed to add alert destination")
			},
		},
		{
			name:        "Test NewStore",
			description: "Test NewStore logic",
			testLogic: func(t *testing.T) {
				am := alert.NewStore()
				assert.NotNil(t, am, "failed to instantiate alert store")
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%s:%d", test.name, i), func(t *testing.T) {
			test.testLogic(t)
		})
	}
}
