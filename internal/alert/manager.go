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

// Manager ... Interface for alert manager
type Manager interface {
	AddSession(core.UUID, *core.AlertPolicy) error
	Transit() chan core.Alert

	core.Subsystem
}

// Config ... Alert manager configuration
// SNSConfig is not part of the RoutingParams as we only support publishing to one SNS client
type Config struct {
	RoutingCfgPath          string
	PagerdutyAlertEventsURL string
	RoutingParams           *core.AlertRoutingParams
	SNSConfig               *client.SNSConfig
}

// alertManager ... Alert manager implementation
type alertManager struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *Config

	store        Store
	interpolator *Interpolator
	cdHandler    CoolDownHandler
	cm           RoutingDirectory

	logger       *zap.Logger
	metrics      metrics.Metricer
	alertTransit chan core.Alert
}

// NewManager ... Instantiates a new alert manager
func NewManager(ctx context.Context, cfg *Config, cm RoutingDirectory) Manager {
	// NOTE - Consider constructing dependencies in higher level
	// abstraction and passing them in

	ctx, cancel := context.WithCancel(ctx)

	// NOTE - Consider adding support for additional sns configurations
	am := &alertManager{
		ctx:          ctx,
		cdHandler:    NewCoolDownHandler(),
		cfg:          cfg,
		cm:           cm,
		cancel:       cancel,
		interpolator: new(Interpolator),
		store:        NewStore(),
		alertTransit: make(chan core.Alert),
		metrics:      metrics.WithContext(ctx),
		logger:       logging.WithContext(ctx),
	}

	return am
}

// AddSession ... Adds a heuristic session to the alert manager store
func (am *alertManager) AddSession(id core.UUID, policy *core.AlertPolicy) error {
	return am.store.AddAlertPolicy(id, policy)
}

// Transit ... Returns inter-subsystem transit channel for receiving alerts
// TODO - Rename this to ingress()
func (am *alertManager) Transit() chan core.Alert {
	return am.alertTransit
}

// handleSlackPost ... Handles posting an alert to slack channels
func (am *alertManager) handleSlackPost(alert core.Alert, policy *core.AlertPolicy) error {
	slackClients := am.cm.GetSlackClients(alert.Sev)
	if slackClients == nil {
		am.logger.Warn("No slack clients defined for criticality", zap.Any("alert", alert))
		return nil
	}

	// Create event trigger
	event := &client.AlertEventTrigger{
		Message: am.interpolator.SlackMessage(alert, policy.Msg),
		Alert:   alert,
	}

	for _, sc := range slackClients {
		resp, err := sc.PostEvent(am.ctx, event)
		if err != nil {
			return err
		}

		if resp.Status != core.SuccessStatus {
			return fmt.Errorf("client %s could not post to slack: %s", sc.GetName(), resp.Message)
		}
		am.logger.Debug("Successfully posted to Slack", zap.String("resp", resp.Message))
		am.metrics.RecordAlertGenerated(alert, core.Slack, sc.GetName())
	}

	return nil
}

// handlePagerDutyPost ... Handles posting an alert to pagerduty
func (am *alertManager) handlePagerDutyPost(alert core.Alert) error {
	pdClients := am.cm.GetPagerDutyClients(alert.Sev)

	if pdClients == nil {
		am.logger.Warn("No pagerduty clients defined for criticality", zap.Any("alert", alert))
		return nil
	}

	event := &client.AlertEventTrigger{
		Message: am.interpolator.PagerDutyMessage(alert),
		Alert:   alert,
	}

	for _, pdc := range pdClients {
		resp, err := pdc.PostEvent(am.ctx, event)
		if err != nil {
			return err
		}

		if resp.Status != core.SuccessStatus {
			return fmt.Errorf("client %s could not post to pagerduty: %s", pdc.GetName(), resp.Message)
		}

		am.logger.Debug("Successfully posted to PagerDuty", zap.Any("resp", resp))
		am.metrics.RecordAlertGenerated(alert, core.PagerDuty, pdc.GetName())
	}

	return nil
}

func (am *alertManager) handleSNSPublish(alert core.Alert, policy *core.AlertPolicy) error {
	event := &client.AlertEventTrigger{
		Message: am.interpolator.SlackMessage(alert, policy.Msg),
		Alert:   alert,
	}

	c := am.cm.GetSNSClient()

	resp, err := c.PostEvent(am.ctx, event)
	if err != nil {
		return err
	}

	if resp.Status != core.SuccessStatus {
		return fmt.Errorf("client %s could not post to sns: %s", c.GetName(), resp.Message)
	}

	am.logger.Debug("Successfully posted to SNS", zap.Any("resp", resp))
	am.metrics.RecordAlertGenerated(alert, core.SNS, c.GetName())
	return nil
}

// EventLoop ... Event loop for alert manager subsystem
func (am *alertManager) EventLoop() error {
	ticker := time.NewTicker(time.Second * 1)

	if am.cfg.RoutingParams == nil {
		am.logger.Warn("No alert routing params defined")
	}

	am.cm.InitializeRouting(am.cfg.RoutingParams)

	for {
		select {
		case <-am.ctx.Done(): // Shutdown
			ticker.Stop()
			return nil

		case <-ticker.C: // Update cool down
			am.cdHandler.Update()

		case alert := <-am.alertTransit: // Upstream alert

			// 1. Fetch alert policy
			policy, err := am.store.GetAlertPolicy(alert.HeuristicID)
			if err != nil {
				am.logger.Error("Could not determine alerting destination", zap.Error(err))
				continue
			}

			// 2. Check if alert is in cool down
			if policy.HasCoolDown() && am.cdHandler.IsCoolDown(alert.HeuristicID) {
				am.logger.Debug("Alert is in cool down",
					zap.String(logging.UUID, alert.HeuristicID.String()))
				continue
			}

			// 3. Log & propagate alert
			am.logger.Info("received alert",
				zap.String(logging.UUID, alert.HeuristicID.String()))

			am.HandleAlert(alert, policy)

			// 4. Add alert to cool down if applicable
			if policy.HasCoolDown() {
				am.cdHandler.Add(alert.HeuristicID, time.Duration(policy.CoolDown)*time.Second)
			}
		}
	}
}

// HandleAlert ... Handles the alert propagation logic
func (am *alertManager) HandleAlert(alert core.Alert, policy *core.AlertPolicy) {
	alert.Sev = policy.Severity()

	if err := am.handleSlackPost(alert, policy); err != nil {
		am.logger.Error("could not post to slack", zap.Error(err))
	}

	if err := am.handlePagerDutyPost(alert); err != nil {
		am.logger.Error("could not post to pagerduty", zap.Error(err))
	}

	if err := am.handleSNSPublish(alert, policy); err != nil {
		am.logger.Error("could not publish to sns", zap.Error(err))
	}
}

// Shutdown ... Shuts down the alert manager subsystem
func (am *alertManager) Shutdown() error {
	am.cancel()
	return nil
}
