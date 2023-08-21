//go:generate mockgen -package mocks --destination ../mocks/client_map.go . ClientMap

package alert

import (
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

type ClientMap interface {
	GetPagerDutyClients(sev core.Severity) []client.PagerDutyClient
	GetSlackClients(sev core.Severity) []client.SlackClient
	InitAlertClients(params *core.AlertRoutingParams)
	SetPagerDutyClients([]client.PagerDutyClient, core.Severity)
	SetSlackClients([]client.SlackClient, core.Severity)
}

type clientMap struct {
	pagerDutyClients map[core.Severity][]client.PagerDutyClient
	slackClients     map[core.Severity][]client.SlackClient
	cfg              *Config
}

func NewClientMap(cfg *Config) ClientMap {
	return &clientMap{
		pagerDutyClients: make(map[core.Severity][]client.PagerDutyClient),
		slackClients:     make(map[core.Severity][]client.SlackClient),
		cfg:              cfg,
	}
}

// GetPagerDutyClients ... Returns the pager duty clients for the given severity level
func (cm *clientMap) GetPagerDutyClients(sev core.Severity) []client.PagerDutyClient {
	return cm.pagerDutyClients[sev]
}

// GetSlackClients ... Returns the slack clients for the given severity level
func (cm *clientMap) GetSlackClients(sev core.Severity) []client.SlackClient {
	return cm.slackClients[sev]
}

// InitAlertClients ... Parses alert routing parameters for each severity level
func (cm *clientMap) InitAlertClients(params *core.AlertRoutingParams) {
	if params != nil {
		cm.alertClientCfgToClientMap(params.AlertRoutes.Low, core.LOW)
		cm.alertClientCfgToClientMap(params.AlertRoutes.Medium, core.MEDIUM)
		cm.alertClientCfgToClientMap(params.AlertRoutes.High, core.HIGH)
	}
}

func (cm *clientMap) SetSlackClients(clients []client.SlackClient, sev core.Severity) {
	cm.slackClients[sev] = clients
}

func (cm *clientMap) SetPagerDutyClients(clients []client.PagerDutyClient, sev core.Severity) {
	cm.pagerDutyClients[sev] = clients
}

// alertClientCfgToClientMap ... Converts alert client config to an alert client map
func (cm *clientMap) alertClientCfgToClientMap(acc *core.AlertClientCfg, sev core.Severity) {
	if acc.Slack != nil {
		for _, cfg := range acc.Slack {
			conf := &client.SlackConfig{
				URL:     cfg.URL,
				Channel: cfg.Channel,
			}
			cli := client.NewSlackClient(conf)
			cm.slackClients[sev] = append(cm.slackClients[sev], cli)
		}
	}

	if acc.PagerDuty != nil {
		for _, cfg := range acc.PagerDuty {
			conf := &client.PagerDutyConfig{
				IntegrationKey: cfg.IntegrationKey,
				AlertEventsURL: cm.cfg.PagerdutyAlertEventsURL,
			}
			cli := client.NewPagerDutyClient(conf)
			cm.pagerDutyClients[sev] = append(cm.pagerDutyClients[sev], cli)
		}
	}
}
