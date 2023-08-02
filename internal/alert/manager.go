//go:generate mockgen -package mocks --destination ../mocks/alert_manager.go --mock_names Manager=AlertManager . Manager

package alert

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"go.uber.org/zap"
)

// Manager ... Interface for alert manager
type Manager interface {
	AddSession(core.SUUID, core.AlertDestination) error
	Transit() chan core.Alert

	core.Subsystem
}

// alertManager ... Alert manager implementation
type alertManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	sc           client.SlackClient
	pdc          client.PagerdutyClient
	store        Store
	interpolator Interpolator

	metrics      metrics.Metricer
	alertTransit chan core.Alert
}

// NewManager ... Instantiates a new alert manager
func NewManager(ctx context.Context, sc client.SlackClient, pdc client.PagerdutyClient) Manager {
	// NOTE - Consider constructing dependencies in higher level
	// abstraction and passing them in

	ctx, cancel := context.WithCancel(ctx)

	am := &alertManager{
		ctx:    ctx,
		cancel: cancel,

		sc:           sc,
		pdc:          pdc,
		interpolator: NewInterpolator(),
		store:        NewStore(),
		alertTransit: make(chan core.Alert),
		metrics:      metrics.WithContext(ctx),
	}

	return am
}

// AddSession ... Adds an heuristic session to the alert manager store
func (am *alertManager) AddSession(sUUID core.SUUID, alertDestination core.AlertDestination) error {
	return am.store.AddAlertDestination(sUUID, alertDestination)
}

// TODO - Rename this to ingress()
// Transit ... Returns inter-subsystem transit channel for receiving alerts
func (am *alertManager) Transit() chan core.Alert {
	return am.alertTransit
}

// handleSlackPost ... Handles posting an alert to slack channel
func (am *alertManager) handleSlackPost(alert core.Alert) error {
	slackMsg := am.interpolator.InterpolateSlackMessage(alert.SUUID, alert.Content)

	resp, err := am.sc.PostData(am.ctx, slackMsg)
	if err != nil {
		return err
	}

	if !resp.Ok && resp.Err != "" {
		return fmt.Errorf(resp.Err)
	}

	return nil
}

// handlePagerdutyPost ... Handles posting an alert to pagerduty
func (am *alertManager) handlePagerdutyPost(alert core.Alert) error {
	pdMsg := am.interpolator.InterpolatePagerdutyMessage(alert.SUUID, alert.Content)
	resp, err := am.pdc.PostEvent(am.ctx, &client.PagerdutyEventTrigger{
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

	return nil
}

// EventLoop ... Event loop for alert manager subsystem
func (am *alertManager) EventLoop() error {
	logger := logging.WithContext(am.ctx)

	for {
		select {
		case <-am.ctx.Done():
			return nil

		case alert := <-am.alertTransit:
			logger.Info("received alert",
				zap.String(logging.SUUIDKey, alert.SUUID.String()))

			alertDest, err := am.store.GetAlertDestination(alert.SUUID)
			if err != nil {
				logger.Error("Could not determine alerting destination", zap.Error(err))
				continue
			}

			alert.Dest = alertDest
			am.metrics.RecordAlertGenerated(alert)

			switch alertDest {
			case core.Slack: // TODO: add more alert destinations
				logger.Debug("Attempting to post alert to slack")

				err := am.handleSlackPost(alert)
				if err != nil {
					logger.Error("Could not post alert to slack", zap.Error(err))
				}

			case core.Pagerduty:
				logger.Debug("Attempting to post alert to pagerduty")
				err := am.handlePagerdutyPost(alert)
				if err != nil {
					logger.Error("Could not post alert to pagerduty", zap.Error(err))
				}

			case core.ThirdParty:
				logger.Error("Attempting to post alert to third_party which is not yet supported")

			default:
				logger.Error("Attempting to post alert to unknown destination",
					zap.String("destination", alertDest.String()))
			}
		}
	}
}

// Shutdown ... Shuts down the alert manager subsystem
func (am *alertManager) Shutdown() error {
	am.cancel()
	return nil
}
