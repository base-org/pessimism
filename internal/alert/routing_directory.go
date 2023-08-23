//go:generate mockgen -package mocks --destination ../mocks/routing_directory.go . RoutingDirectory

package alert

import (
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

type RoutingDirectory interface {
	GetPagerDutyClients(sev core.Severity) []client.PagerDutyClient
	GetSlackClients(sev core.Severity) []client.SlackClient
	InitializeRouting(params *core.AlertRoutingParams)
	SetPagerDutyClients([]client.PagerDutyClient, core.Severity)
	SetSlackClients([]client.SlackClient, core.Severity)
}

type routingDirectory struct {
	pagerDutyClients map[core.Severity][]client.PagerDutyClient
	slackClients     map[core.Severity][]client.SlackClient
	cfg              *Config
}

func NewRoutingDirectory(cfg *Config) RoutingDirectory {
	return &routingDirectory{
		cfg:              cfg,
		pagerDutyClients: make(map[core.Severity][]client.PagerDutyClient),
		slackClients:     make(map[core.Severity][]client.SlackClient),
	}
}

// GetPagerDutyClients ... Returns the pager duty clients for the given severity level
func (rd *routingDirectory) GetPagerDutyClients(sev core.Severity) []client.PagerDutyClient {
	return rd.pagerDutyClients[sev]
}

// GetSlackClients ... Returns the slack clients for the given severity level
func (rd *routingDirectory) GetSlackClients(sev core.Severity) []client.SlackClient {
	return rd.slackClients[sev]
}

func (rd *routingDirectory) SetSlackClients(clients []client.SlackClient, sev core.Severity) {
	copy(rd.slackClients[sev][0:], clients)
}

func (rd *routingDirectory) SetPagerDutyClients(clients []client.PagerDutyClient, sev core.Severity) {
	copy(rd.pagerDutyClients[sev][0:], clients)
}

// InitializeRouting ... Parses alert routing parameters for each severity level
func (rd *routingDirectory) InitializeRouting(params *core.AlertRoutingParams) {
	if params != nil {
		rd.paramsToRouteDirectory(params.AlertRoutes.Low, core.LOW)
		rd.paramsToRouteDirectory(params.AlertRoutes.Medium, core.MEDIUM)
		rd.paramsToRouteDirectory(params.AlertRoutes.High, core.HIGH)
	}
}

// paramsToRouteDirectory ... Converts alert client config to an alert client map
func (rd *routingDirectory) paramsToRouteDirectory(acc *core.AlertClientCfg, sev core.Severity) {
	if acc == nil {
		return
	}

	if acc.Slack != nil {
		for _, cfg := range acc.Slack {
			conf := &client.SlackConfig{
				URL:     cfg.URL,
				Channel: cfg.Channel,
			}
			client := client.NewSlackClient(conf)
			rd.slackClients[sev] = append(rd.slackClients[sev], client)
		}
	}

	if acc.PagerDuty != nil {
		for _, cfg := range acc.PagerDuty {
			conf := &client.PagerDutyConfig{
				IntegrationKey: cfg.IntegrationKey,
				AlertEventsURL: rd.cfg.PagerdutyAlertEventsURL,
			}
			client := client.NewPagerDutyClient(conf)
			rd.pagerDutyClients[sev] = append(rd.pagerDutyClients[sev], client)
		}
	}
}
