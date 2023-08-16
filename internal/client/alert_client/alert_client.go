package alert_client

import (
	"context"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

// AlertStatus ... A standardized response status for alert clients
type AlertStatus string

const (
	SuccessStatus AlertStatus = "success"
	FailureStatus AlertStatus = "failure"
)

// AlertClient ... An interface for alert clients to implement
type AlertClient interface {
	PostEvent(ctx context.Context, data *AlertEventTrigger) (*AlertAPIResponse, error)
}

// AlertEventTrigger ... A standardized event trigger for alert clients
type AlertEventTrigger struct {
	Message  string
	Severity core.Severity
	DedupKey core.PUUID
}

// AlertAPIResponse ... A standardized response for alert clients
type AlertAPIResponse struct {
	Status  AlertStatus
	Message string
}

// ToPagerdutyEvent ... Converts an AlertEventTrigger to a PagerDutyEventTrigger
func (a *AlertEventTrigger) ToPagerdutyEvent() *PagerDutyEventTrigger {
	return &PagerDutyEventTrigger{
		DedupKey: a.DedupKey.String(),
		Severity: a.Severity.ToPagerdutySev(),
		Message:  a.Message,
	}
}

// Config ... A global config to be used to instantiate alert clients
type Config struct {
	PagerDutyEventUrl string
	RouteMapCfgPath   string
}

// AlertClientMap ... A map for alert clients
type AlertClientMap struct {
	SlackClients     map[string][]AlertClient
	PagerdutyClients map[string][]AlertClient
	cfg              *Config
	Params           *core.AlertRoutesTable
}

func NewAlertClientMap(cfg *Config, params *core.AlertRoutesTable) *AlertClientMap {
	return &AlertClientMap{
		cfg:              cfg,
		Params:           params,
		PagerdutyClients: make(map[string][]AlertClient),
		SlackClients:     make(map[string][]AlertClient),
	}
}

// ParseCfgToRouteMap TODO this function is a mess but it works for now
// ParseCfgToRouteMap ... Returns a mapping of alert clients
func (a *AlertClientMap) ParseCfgToRouteMap(routes ...core.AlertRoute) *AlertClientMap {

	// Loop through alert routes table
	for k, v := range a.Params.AlertRoutes {
		for _, route := range routes {
			for _, n := range v[string(route)] {
				for _, cfg := range n {

					// 3. Switch on route to determine which client to instantiate
					switch route {
					case core.AlertRoutePagerDuty:
						pdcfg := &PagerDutyConfig{
							Priority:       k,
							IntegrationKey: cfg.IntegrationKey,
							AlertEventsURL: a.cfg.PagerDutyEventUrl,
						}
						cli := NewPagerDutyClient(pdcfg)
						a.PagerdutyClients[k] = append(a.PagerdutyClients[k], cli)
					case core.AlertRouteSlack:
						scfg := &SlackConfig{
							Channel:  cfg.Channel,
							URL:      cfg.URL,
							Priority: k,
						}
						cli := NewSlackClient(scfg)
						a.SlackClients[k] = append(a.SlackClients[k], cli)
					// If route is not supported, log a warning
					default:
						logging.NoContext().Warn("Invalid alert route provided")
					}
				}
			}
		}
	}

	return a
}
