//go:generate mockgen -package mocks --destination ../mocks/alert_manager.go --mock_names Manager=AlertManager . Manager

package alert

import (
	"context"
	"fmt"
	"github.com/base-org/pessimism/internal/client/alert_client"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

var SupportedAlertClients = []core.AlertRoute{"slack", "pagerduty"}

//// NOTE - This should be user defined in the future
//// with modularity in mind so that users can define
//// their own independent alerting policies
//func getSevMap() map[core.Severity][]core.AlertDestination {
//	return map[core.Severity][]core.AlertDestination{
//		core.UNKNOWN: {core.AlertDestination(0)},
//		core.LOW:     {core.Slack},
//		core.MEDIUM:  {core.PagerDuty, core.Slack},
//		core.HIGH:    {core.PagerDuty, core.Slack},
//	}
//}

// Manager ... Interface for alert manager
type Manager interface {
	AddSession(core.SUUID, *core.AlertPolicy) error
	Transit() chan core.Alert

	core.Subsystem
}

// Config ... Alert manager configuration
type Config struct {
	AlertRoutingCfgPath     string
	SlackURL                string
	PagerdutyAlertEventsUrl string
}

// alertManager ... Alert manager implementation
type alertManager struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *Config

	store        Store
	interpolator Interpolator
	cdHandler    CoolDownHandler
	acm          *alert_client.AlertClientMap

	logger       *zap.Logger
	metrics      metrics.Metricer
	alertTransit chan core.Alert
}

// NewManager ... Instantiates a new alert manager
func NewManager(ctx context.Context, acm *alert_client.AlertClientMap) Manager {
	// NOTE - Consider constructing dependencies in higher level
	// abstraction and passing them in

	ctx, cancel := context.WithCancel(ctx)

	am := &alertManager{
		ctx:       ctx,
		cdHandler: NewCoolDownHandler(),
		acm:       acm.ParseCfgToRouteMap(SupportedAlertClients...),

		cancel:       cancel,
		interpolator: NewInterpolator(),
		store:        NewStore(),
		alertTransit: make(chan core.Alert),
		metrics:      metrics.WithContext(ctx),
		logger:       logging.WithContext(ctx),
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

// handleSlackPost ... Handles posting an alert to slack channels
func (am *alertManager) handleSlackPost(alert core.Alert, policy *core.AlertPolicy) error {
	// Check if user has defined slack clients for this criticality
	if _, ok := am.acm.SlackClients[alert.Criticality.String()]; !ok {
		return fmt.Errorf("no slack clients for criticality: %s", alert.Criticality.String())
	}

	// Create event trigger
	event := &alert_client.AlertEventTrigger{
		Message:  am.interpolator.InterpolateSlackMessage(alert.SUUID, alert.Content, policy.Msg),
		Severity: alert.Criticality,
	}

	for _, sc := range am.acm.SlackClients[alert.Criticality.String()] {
		resp, err := sc.PostEvent(am.ctx, event)
		if err != nil {
			return err
		}

		if resp.Status != alert_client.SuccessStatus {
			return fmt.Errorf("could not post to slack: %s", resp.Message)
		}
	}

	return nil
}

// handlePagerDutyPost ... Handles posting an alert to pagerduty
func (am *alertManager) handlePagerDutyPost(alert core.Alert) error {
	// Check if pagerduty client array exists
	if _, ok := am.acm.PagerdutyClients[alert.Criticality.String()]; !ok {
		return fmt.Errorf("no slack clients for criticality: %s", alert.Criticality.String())
	}

	pdMsg := am.interpolator.InterpolatePagerDutyMessage(alert.SUUID, alert.Content)

	event := &alert_client.AlertEventTrigger{
		Message:  pdMsg,
		DedupKey: alert.PUUID,
		Severity: alert.Criticality,
	}

	for _, pdc := range am.acm.PagerdutyClients[alert.Criticality.String()] {
		resp, err := pdc.PostEvent(am.ctx, event)
		if err != nil {
			return err
		}

		if resp.Status != alert_client.SuccessStatus {
			return fmt.Errorf("could not post to pagerduty: %s", resp.Message)
		}
	}

	return nil
}

// EventLoop ... Event loop for alert manager subsystem
func (am *alertManager) EventLoop() error {
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
				am.logger.Error("Could not determine alerting destination", zap.Error(err))
				continue
			}

			// 2. Check if alert is in cool down
			if policy.HasCoolDown() && am.cdHandler.IsCoolDown(alert.SUUID) {
				am.logger.Debug("Alert is in cool down",
					zap.String(logging.SUUIDKey, alert.SUUID.String()))
				continue
			}

			// 3. Log & propagate alert
			am.logger.Info("received alert",
				zap.String(logging.SUUIDKey, alert.SUUID.String()))

			am.HandleAlert(alert, policy)

			// 4. Add alert to cool down if applicable
			if policy.HasCoolDown() {
				am.cdHandler.Add(alert.SUUID, time.Duration(policy.CoolDown)*time.Second)
			}
		}
	}
}

// HandleAlert ... Handles the alert propagation logic
func (am *alertManager) HandleAlert(alert core.Alert, policy *core.AlertPolicy) {
	alert.Criticality = policy.Severity()

	if err := am.handleSlackPost(alert, policy); err != nil {
		am.logger.Error("could not post to slack", zap.Error(err))
	}

	if err := am.handlePagerDutyPost(alert); err != nil {
		am.logger.Error("could not post to pagerduty", zap.Error(err))
	}
}

// Shutdown ... Shuts down the alert manager subsystem
func (am *alertManager) Shutdown() error {
	am.cancel()
	return nil
}
