package alert_test

import (
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func getCfg() *config.Config {
	return &config.Config{
		AlertConfig: &alert.Config{
			SNSConfig: &client.SNSConfig{
				TopicArn: "test",
			},
			RoutingParams: &core.AlertRoutingParams{
				AlertRoutes: &core.SeverityMap{
					Low: &core.AlertClientCfg{
						Slack: map[string]*core.AlertConfig{
							"test1": {
								Channel: "test1",
								URL:     "test1",
							},
						},
					},
					Medium: &core.AlertClientCfg{
						PagerDuty: map[string]*core.AlertConfig{
							"test1": {
								IntegrationKey: "test1",
							},
						},
						Slack: map[string]*core.AlertConfig{
							"test2": {
								Channel: "test2",
								URL:     "test2",
							},
						},
					},
					High: &core.AlertClientCfg{
						PagerDuty: map[string]*core.AlertConfig{
							"test1": {
								IntegrationKey: "test1",
							},
							"test2": {
								IntegrationKey: "test2",
							},
						},
						Slack: map[string]*core.AlertConfig{
							"test2": {
								Channel: "test2",
								URL:     "test2",
							},
							"test3": {
								Channel: "test3",
								URL:     "test3",
							},
						},
					},
				},
			},
		},
	}
}

func Test_AlertClientCfgToClientMap(t *testing.T) {
	tests := []struct {
		name        string
		description string
		testLogic   func(t *testing.T)
	}{
		{
			name:        "Test AlertClientCfgToClientMap Success",
			description: "Test AlertClientCfgToClientMap successfully creates alert clients",
			testLogic: func(t *testing.T) {
				cfg := getCfg()

				cm := alert.NewRoutingDirectory(cfg.AlertConfig)

				assert.NotNil(t, cm, "client map is nil")
				cm.InitializeRouting(cfg.AlertConfig.RoutingParams)

				assert.Len(t, cm.GetSlackClients(core.LOW), 1)
				assert.Len(t, cm.GetPagerDutyClients(core.LOW), 0)
				assert.Len(t, cm.GetSlackClients(core.MEDIUM), 1)
				assert.Len(t, cm.GetPagerDutyClients(core.MEDIUM), 1)
				assert.Len(t, cm.GetSlackClients(core.HIGH), 2)
				assert.Len(t, cm.GetPagerDutyClients(core.HIGH), 2)
			},
		},
		{
			name:        "Test AlertClientCfgToClientMap Pagerduty Nil",
			description: "Test AlertClientCfgToClientMap doesn't fail when pagerduty is nil",
			testLogic: func(t *testing.T) {
				cfg := getCfg()
				cfg.AlertConfig.RoutingParams.AlertRoutes.Medium.PagerDuty = nil
				cm := alert.NewRoutingDirectory(cfg.AlertConfig)
				assert.NotNil(t, cm, "client map is nil")

				cm.InitializeRouting(cfg.AlertConfig.RoutingParams)
				assert.Len(t, cm.GetSlackClients(core.LOW), 1)
				assert.Len(t, cm.GetPagerDutyClients(core.LOW), 0)
				assert.Len(t, cm.GetSlackClients(core.MEDIUM), 1)
				assert.Len(t, cm.GetPagerDutyClients(core.MEDIUM), 0)
				assert.Len(t, cm.GetSlackClients(core.HIGH), 2)
				assert.Len(t, cm.GetPagerDutyClients(core.HIGH), 2)
			},
		},
		{
			name:        "Test AlertClientCfgToClientMap Nil Slack",
			description: "Test AlertClientCfgToClientMap doesn't fail when slack is nil",
			testLogic: func(t *testing.T) {
				cfg := getCfg()
				cfg.AlertConfig.RoutingParams.AlertRoutes.Medium.Slack = nil
				cm := alert.NewRoutingDirectory(cfg.AlertConfig)
				assert.NotNil(t, cm, "client map is nil")

				cm.InitializeRouting(cfg.AlertConfig.RoutingParams)
				assert.Len(t, cm.GetSlackClients(core.LOW), 1)
				assert.Len(t, cm.GetPagerDutyClients(core.LOW), 0)
				assert.Len(t, cm.GetSlackClients(core.MEDIUM), 0)
				assert.Len(t, cm.GetPagerDutyClients(core.MEDIUM), 1)
				assert.Len(t, cm.GetSlackClients(core.HIGH), 2)
				assert.Len(t, cm.GetPagerDutyClients(core.HIGH), 2)
			},
		},
		{
			name:        "Test AlertClientCfgToClientMap Nil Params",
			description: "Test AlertClientCfgToClientMap doesn't fail when params are nil",
			testLogic: func(t *testing.T) {
				cfg := getCfg()

				cfg.AlertConfig.RoutingParams = nil

				cm := alert.NewRoutingDirectory(cfg.AlertConfig)
				assert.NotNil(t, cm, "client map is nil")

				cm.InitializeRouting(cfg.AlertConfig.RoutingParams)

				assert.Len(t, cm.GetSlackClients(core.LOW), 0)
				assert.Len(t, cm.GetPagerDutyClients(core.LOW), 0)
				assert.Len(t, cm.GetSlackClients(core.MEDIUM), 0)
				assert.Len(t, cm.GetPagerDutyClients(core.MEDIUM), 0)
				assert.Len(t, cm.GetSlackClients(core.HIGH), 0)
				assert.Len(t, cm.GetPagerDutyClients(core.HIGH), 0)
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d:%s", i, test.name), func(t *testing.T) {
			test.testLogic(t)
		})
	}

}
