//go:generate mockgen -package mocks --destination ../mocks/routing_directory.go . RoutingDirectory

package alert

import (
	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

// RoutingDirectory ... Interface for routing directory
type RoutingDirectory interface {
	GetPagerDutyClients(sev core.Severity) []client.PagerDutyClient
	GetSlackClients(sev core.Severity) []client.SlackClient
	GetTelegramClients(sev core.Severity) []client.TelegramClient
	InitializeRouting(params *core.AlertRoutingParams)
	SetPagerDutyClients([]client.PagerDutyClient, core.Severity)
	SetSlackClients([]client.SlackClient, core.Severity)
	GetSNSClient() client.SNSClient
	SetSNSClient(client.SNSClient)
	SetTelegramClients([]client.TelegramClient, core.Severity)
}

// routingDirectory ... Routing directory implementation
// NOTE: This implementation works for now, but if we add more routing clients in the future,
// we should consider refactoring this to be more generic
// Only one SNS client is needed in most cases. If we need to support multiple SNS clients, we can refactor this
type routingDirectory struct {
	pagerDutyClients map[core.Severity][]client.PagerDutyClient
	slackClients     map[core.Severity][]client.SlackClient
	snsClient        client.SNSClient
	telegramClients  map[core.Severity][]client.TelegramClient
	cfg              *Config
}

// NewRoutingDirectory ... Instantiates a new routing directory
func NewRoutingDirectory(cfg *Config) RoutingDirectory {
	return &routingDirectory{
		cfg:              cfg,
		pagerDutyClients: make(map[core.Severity][]client.PagerDutyClient),
		slackClients:     make(map[core.Severity][]client.SlackClient),
		snsClient:        nil,
		telegramClients:  make(map[core.Severity][]client.TelegramClient),
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

// GetTelegramClients ... Returns the telegram clients for the given severity level
func (rd *routingDirectory) GetTelegramClients(sev core.Severity) []client.TelegramClient {
	return rd.telegramClients[sev]
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

// SetTelegramClients ... Sets the telegram clients for the given severity level
func (rd *routingDirectory) SetTelegramClients(clients []client.TelegramClient, sev core.Severity) {
	rd.telegramClients[sev] = make([]client.TelegramClient, len(clients))
	copy(rd.telegramClients[sev], clients)
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

	if acc.Telegram != nil {
		for name, cfg := range acc.Telegram {
			conf := &client.TelegramConfig{
				Token:  cfg.Token.String(),
				ChatID: cfg.ChatID.String(),
			}
			client, err := client.NewTelegramClient(conf, name)
			if err != nil {
				logging.NoContext().Error("Failed to create Telegram client", zap.String("name", name), zap.Error(err))
				continue
			}
			rd.telegramClients[sev] = append(rd.telegramClients[sev], client)
		}
	}
}
