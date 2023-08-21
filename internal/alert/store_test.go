package alert_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"

	"github.com/stretchr/testify/assert"
)

func Test_Store(t *testing.T) {

	cfg := &config.Config{
		AlertConfig: &alert.Config{
			AlertRoutingCfgPath:     "test_data/alert-routing-test.yaml",
			PagerdutyAlertEventsURL: "test",
		},
	}

	err := cfg.ParseAlertConfig()
	assert.Nil(t, err, fmt.Sprintf("failed to parse alert config: %v", err))

	var tests = []struct {
		name        string
		description string
		testLogic   func(t *testing.T)
	}{
		{
			name:        "Test Get Alert Policy Success",
			description: "Test GetAlertPolicy",
			testLogic: func(t *testing.T) {
				am := alert.NewStore(cfg.AlertConfig)

				sUUID := core.MakeSUUID(core.Layer1, core.Live, core.BalanceEnforcement)
				policy := &core.AlertPolicy{
					Msg:  "test message",
					Dest: core.Slack.String(),
				}

				err := am.AddAlertPolicy(sUUID, policy)
				assert.NoError(t, err, "failed to add Alert Policy")

				actualPolicy, err := am.GetAlertPolicy(sUUID)
				assert.NoError(t, err, "failed to get Alert Policy")
				assert.Equal(t, policy, actualPolicy, "Alert Policy mismatch")
			},
		},
		{
			name:        "Test Add Alert Policy Success",
			description: "Test adding of arbitrary Alert Policies",
			testLogic: func(t *testing.T) {
				am := alert.NewStore(cfg.AlertConfig)

				sUUID := core.MakeSUUID(core.Layer1, core.Live, core.BalanceEnforcement)
				policy := &core.AlertPolicy{
					Dest: core.Slack.String(),
				}

				err := am.AddAlertPolicy(sUUID, policy)
				assert.NoError(t, err, "failed to add Alert Policy")

				// add again
				err = am.AddAlertPolicy(sUUID, policy)
				assert.Error(t, err, "failed to add Alert Policy")
			},
		},
		{
			name:        "Test NewStore",
			description: "Test NewStore logic",
			testLogic: func(t *testing.T) {
				am := alert.NewStore(cfg.AlertConfig)
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
