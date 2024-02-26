//go:generate mockgen -package mocks --destination ../mocks/routing_directory.go . RoutingDirectory

package alert

import (
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

// RoutingDirectory ... Interface for routing directory
type RoutingDirectory interface {
	GetPagerDutyClients(sev core.Severity) []client.PagerDutyClient
	GetSlackClients(sev core.Severity) []client.SlackClient
	InitializeRouting(params *core.AlertRoutingParams)
	SetPagerDutyClients([]client.PagerDutyClient, core.Severity)
	SetSlackClients([]client.SlackClient, core.Severity)
	GetSNSClient() client.SNSClient
	SetSNSClient(client.SNSClient)
}

// routingDirectory ... Routing directory implementation
// NOTE: This implementation works for now, but if we add more routing clients in the future,
// we should consider refactoring this to be more generic
// Only one SNS client is needed in most cases. If we need to support multiple SNS clients, we can refactor this
type routingDirectory struct {
	pagerDutyClients map[core.Severity][]client.PagerDutyClient
	slackClients     map[core.Severity][]client.SlackClient
	snsClient        client.SNSClient
	cfg              *Config
}

// NewRoutingDirectory ... Instantiates a new routing directory
func NewRoutingDirectory(cfg *Config) RoutingDirectory {
	return &routingDirectory{
		cfg:              cfg,
		pagerDutyClients: make(map[core.Severity][]client.PagerDutyClient),
		slackClients:     make(map[core.Severity][]client.SlackClient),
		snsClient:        nil,
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

// SetSlackClients ... Sets the slack clients for the given severity level
func (rd *routingDirectory) SetSlackClients(clients []client.SlackClient, sev core.Severity) {
	copy(rd.slackClients[sev][0:], clients)
}

func (rd *routingDirectory) GetSNSClient() client.SNSClient {
	return rd.snsClient
}

func (rd *routingDirectory) SetSNSClient(client client.SNSClient) {
	rd.snsClient = client
}

// SetPagerDutyClients ... Sets the pager duty clients for the given severity level
func (rd *routingDirectory) SetPagerDutyClients(clients []client.PagerDutyClient, sev core.Severity) {
	copy(rd.pagerDutyClients[sev][0:], clients)
}

// InitializeRouting ... Parses alert routing parameters for each severity level
func (rd *routingDirectory) InitializeRouting(params *core.AlertRoutingParams) {
	rd.snsClient = client.NewSNSClient(rd.cfg.SNSConfig, "sns")
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
		for name, cfg := range acc.Slack {
			conf := &client.SlackConfig{
				URL:     cfg.URL.String(),
				Channel: cfg.Channel.String(),
			}
			client := client.NewSlackClient(conf, name)
			rd.slackClients[sev] = append(rd.slackClients[sev], client)
		}
	}

	if acc.PagerDuty != nil {
		for name, cfg := range acc.PagerDuty {
			conf := &client.PagerDutyConfig{
				IntegrationKey: cfg.IntegrationKey.String(),
				AlertEventsURL: rd.cfg.PagerdutyAlertEventsURL,
			}
			client := client.NewPagerDutyClient(conf, name)
			rd.pagerDutyClients[sev] = append(rd.pagerDutyClients[sev], client)
		}
	}
}
