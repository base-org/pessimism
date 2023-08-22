package alert_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_AlertClientCfgToClientMap(t *testing.T) {

	cfg := config.Config{
		AlertConfig: &alert.Config{
			RoutingCfgPath:          "test_data/alert-routing-test.yaml",
			PagerdutyAlertEventsURL: "test",
		},
	}

	cm := alert.NewClientMap(cfg.AlertConfig)

	err := cfg.ParseAlertConfig()
	assert.Nil(t, err, fmt.Sprintf("failed to parse alert config: %v", err))

	assert.NotNil(t, cm, "client map is nil")
	cm.InitAlertClients(cfg.AlertConfig.AlertRoutingParams)
	assert.Len(t, cm.GetSlackClients(core.LOW), 1)
	assert.Len(t, cm.GetPagerDutyClients(core.LOW), 0)
	assert.Len(t, cm.GetSlackClients(core.MEDIUM), 1)
	assert.Len(t, cm.GetPagerDutyClients(core.MEDIUM), 1)
	assert.Len(t, cm.GetSlackClients(core.HIGH), 2)
	assert.Len(t, cm.GetPagerDutyClients(core.HIGH), 2)
}
