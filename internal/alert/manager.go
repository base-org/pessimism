//go:generate mockgen -package mocks --destination ../mocks/alert_manager.go --mock_names Manager=AlertManager . Manager

package alert

import (
	"context"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

// NOTE - This should be user defined in the future
// with modularity in mind so that users can define
// their own independent alerting policies
func getSevMap() map[core.Severity][]core.AlertDestination {
	return map[core.Severity][]core.AlertDestination{
		core.UNKNOWN: {core.AlertDestination(0)},
		core.LOW:     {core.Slack},
		core.MEDIUM:  {core.PagerDuty, core.Slack},
		core.HIGH:    {core.PagerDuty, core.Slack},
	}
}

// Manager ... Interface for alert manager
type Manager interface {
	AddSession(core.SUUID, *core.AlertPolicy) error
	Transit() chan core.Alert

	core.Subsystem
}

// Config ... Alert manager configuration
type Config struct {
	SlackConfig        *client.SlackConfig
	MediumPagerDutyCfg *client.PagerDutyConfig
	HighPagerDutyCfg   *client.PagerDutyConfig
}

// alertManager ... Alert manager implementation
type alertManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	store        Store
	interpolator Interpolator
	cdHandler    CoolDownHandler

	metrics      metrics.Metricer
	alertTransit chan core.Alert
	pdcP0        client.PagerDutyClient
	pdcP1        client.PagerDutyClient
	sc           client.SlackClient
}

// NewManager ... Instantiates a new alert manager
func NewManager(ctx context.Context, sc client.SlackClient, pdc client.PagerDutyClient,
	hpdc client.PagerDutyClient) Manager {
	// NOTE - Consider constructing dependencies in higher level
	// abstraction and passing them in

	ctx, cancel := context.WithCancel(ctx)

	am := &alertManager{
		ctx:       ctx,
		cdHandler: NewCoolDownHandler(),
		sc:        sc,

		// NOTE - This is a major regression and is a quick hack to enable
		// multi-service paging in the short-term. Going forward, all alerting
		// configurations should be highly configurable and non-opinionated.
		pdcP0: hpdc,
		pdcP1: pdc,

		cancel:       cancel,
		interpolator: NewInterpolator(),
		store:        NewStore(),
		alertTransit: make(chan core.Alert),
		metrics:      metrics.WithContext(ctx),
	}

	return am
}

// AddSession ... Adds a heuristic session to the alert manager store
func (am *alertManager) AddSession(sUUID core.SUUID, policy *core.AlertPolicy) error {
	return am.store.AddAlertPolicy(sUUID, policy)
}

// TODO - Rename this to ingress()
// Transit ... Returns inter-subsystem transit channel for receiving alerts
func (am *alertManager) Transit() chan core.Alert {
	return am.alertTransit
}

// handleSlackPost ... Handles posting an alert to slack channel
func (am *alertManager) handleSlackPost(sUUID core.SUUID, content string, msg string) error {
	slackMsg := am.interpolator.InterpolateSlackMessage(sUUID, content, msg)

	resp, err := am.sc.PostData(am.ctx, slackMsg)
	if err != nil {
		return err
	}

	if !resp.Ok && resp.Err != "" {
		return fmt.Errorf(resp.Err)
	}

	return nil
}

// handlePagerDutyPost ... Handles posting an alert to pagerduty
func (am *alertManager) handlePagerDutyPost(alert core.Alert) error {
	clients := []client.PagerDutyClient{am.pdcP1}
	if alert.Criticality == core.HIGH {
		clients = append(clients, am.pdcP0)
	}

	pdMsg := am.interpolator.InterpolatePagerDutyMessage(alert.SUUID, alert.Content)

	for _, pdc := range clients {
		resp, err := pdc.PostEvent(am.ctx, &client.PagerDutyEventTrigger{
			Message:  pdMsg,
			Action:   client.Trigger,
			Severity: client.Critical,
			DedupKey: alert.SUUID.String(),
		})
		if err != nil {
			return err
		}

		if resp.Status != string(client.SuccessStatus) {
			return fmt.Errorf("could not post to pagerduty: %s", resp.Status)
		}
	}

	return nil
}

// EventLoop ... Event loop for alert manager subsystem
func (am *alertManager) EventLoop() error {
	logger := logging.WithContext(am.ctx)
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-am.ctx.Done(): // Shutdown
			ticker.Stop()
			return nil

		case <-ticker.C: // Update cool down
			am.cdHandler.Update()

		case alert := <-am.alertTransit: // Upstream alert

			// 1. Fetch alert policy
			policy, err := am.store.GetAlertPolicy(alert.SUUID)
			if err != nil {
				logger.Error("Could not determine alerting destination", zap.Error(err))
				continue
			}

			// 2. Check if alert is in cool down
			if policy.HasCoolDown() && am.cdHandler.IsCoolDown(alert.SUUID) {
				logger.Debug("Alert is in cool down",
					zap.String(logging.SUUIDKey, alert.SUUID.String()))
				continue
			}

			// 3. Log & propagate alert
			logger.Info("received alert",
				zap.String(logging.SUUIDKey, alert.SUUID.String()))

			am.HandleAlert(alert, policy)
			am.metrics.RecordAlertGenerated(alert)

			// 4. Add alert to cool down if applicable
			if policy.HasCoolDown() {
				am.cdHandler.Add(alert.SUUID, time.Duration(policy.CoolDown)*time.Second)
			}
		}
	}
}

// HandleAlert ... Handles the alert propagation logic
func (am *alertManager) HandleAlert(alert core.Alert, policy *core.AlertPolicy) {
	logger := logging.WithContext(am.ctx)

	logger.Debug("Attempting to post alert to slack")
	err := am.handleSlackPost(alert.SUUID, alert.Content, policy.Message())
	if err != nil {
		logger.Error("Could not post alert to slack", zap.Error(err))
		locations := []core.AlertDestination{policy.Destination()}

		am.metrics.RecordAlertGenerated(alert)
		alert.Criticality = policy.Severity()

		// Fetch alerting destinations if severity is provided
		if policy.Severity() != core.UNKNOWN {
			locations = getSevMap()[policy.Severity()]
		}

		// Iterate over alerting destinations and propagate alert
		for _, dest := range locations {
			am.propagate(dest, alert, policy)
		}
	}

}

// propagate ... Propagates an alert to a destination
func (am *alertManager) propagate(dest core.AlertDestination, alert core.Alert, policy *core.AlertPolicy) {
	logger := logging.WithContext(am.ctx)

	switch dest {
	case core.Slack: // TODO: add more alert destinations
		logger.Debug("Attempting to post alert to slack")

		err := am.handleSlackPost(alert.SUUID, alert.Content, policy.Message())
		if err != nil {
			logger.Error("Could not post alert to slack", zap.Error(err))
		}

	case core.PagerDuty:
		logger.Debug("Attempting to post alert to pagerduty")
		err := am.handlePagerDutyPost(alert)
		if err != nil {
			logger.Error("Could not post alert to pagerduty", zap.Error(err))
		}

	case core.ThirdParty:
		logger.Error("Attempting to post alert to third_party which is not yet supported")

	default:
		logger.Error("Attempting to post alert to unknown destination",
			zap.String("destination", policy.Destination().String()))
	}
}

// Shutdown ... Shuts down the alert manager subsystem
func (am *alertManager) Shutdown() error {
	am.cancel()
	return nil
}
