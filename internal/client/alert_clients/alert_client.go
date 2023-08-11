package alert_clients

import (
	"context"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

type AlertStatus string

const (
	SuccessStatus AlertStatus = "success"
	FailureStatus AlertStatus = "failure"
)

type AlertClient interface {
	PostEvent(ctx context.Context, data *AlertEventTrigger) (*AlertAPIResponse, error)
}

type AlertEventTrigger struct {
	Message  string
	Severity core.Severity
	DedupKey core.PUUID
}

type AlertClientMap struct {
	SlackClients     map[string][]AlertClient
	PagerdutyClients map[string][]AlertClient
}

type AlertAPIResponse struct {
	Status  AlertStatus
	Message string
}

func (a *AlertEventTrigger) ToPagerdutyEvent() *PagerDutyEventTrigger {
	return &PagerDutyEventTrigger{
		DedupKey: a.DedupKey.String(),
		Severity: a.Severity.ToPagerdutySev(),
		Message:  a.Message,
	}
}

// GetClientMap TODO this function is a mess but it works for now
// GetClientMap ... Returns a mapping of alert clients
func GetClientMap(a *core.AlertRoutesTable, routes ...core.AlertRoute) *AlertClientMap {

	// 1. Create a new alert client map
	acm := &AlertClientMap{
		PagerdutyClients: make(map[string][]AlertClient),
		SlackClients:     make(map[string][]AlertClient),
	}

	// 2. Loop through alert routes table
	for k, v := range a.AlertRoutes {
		for _, route := range routes {
			for _, n := range v[string(route)] {
				for _, cfg := range n {

					// 3. Switch on route to determine which client to instantiate
					switch route {
					case core.AlertRoutePagerDuty:
						pdcfg := &PagerDutyConfig{
							Priority:       k,
							IntegrationKey: cfg.IntegrationKey,
						}
						cli := NewPagerDutyClient(pdcfg)
						acm.PagerdutyClients[k] = append(acm.PagerdutyClients[k], cli)
					case core.AlertRouteSlack:
						scfg := &SlackConfig{
							Channel:  cfg.Channel,
							URL:      cfg.URL,
							Priority: k,
						}
						cli := NewSlackClient(scfg)
						acm.SlackClients[k] = append(acm.SlackClients[k], cli)
					// If route is not supported, log a warning
					default:
						logging.NoContext().Warn("Invalid alert route provided")
					}
				}
			}
		}
	}

	return acm
}
