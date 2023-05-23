//go:generate mockgen -package mocks --destination ../mocks/alert_manager.go . AlertingManager

package alert

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// AlertingManager ... Interface for alert manager
type AlertingManager interface {
	AddInvariantSession(core.InvSessionUUID, core.AlertDestination) error
	EventLoop(ctx context.Context) error
	Transit() chan core.Alert
}

// alertManager ... Alert manager implementation
type alertManager struct {
	sc                  client.SlackClient
	invariantAlertStore AlertStore
	interpolator        Interpolator

	alertTransit chan core.Alert
}

// NewManager ... Instantiates a new alert manager
func NewManager(sc client.SlackClient) (AlertingManager, func()) {
	am := &alertManager{
		sc:                  sc,
		interpolator:        NewInterpolator(),
		invariantAlertStore: NewAlertStore(),
		alertTransit:        make(chan core.Alert, 0),
	}

	shutDown := func() {
		close(am.alertTransit)
	}

	return am, shutDown
}

// AddInvariantSession ... Adds an invariant session to the alert manager store
func (am *alertManager) AddInvariantSession(sUUID core.InvSessionUUID, alertDestination core.AlertDestination) error {
	return am.invariantAlertStore.AddAlertDestination(sUUID, alertDestination)
}

// Transit ... Returns inter-subsystem transit channel for receiving alerts
func (am *alertManager) Transit() chan core.Alert {
	return am.alertTransit
}

// handleSlackPost ... Handles posting an alert to slack channel
func (am *alertManager) handleSlackPost(alert core.Alert) error {

	slackMsg := am.interpolator.InterpolateSlackMessage(alert.SUUID, alert.Content)

	resp, err := am.sc.PostData(slackMsg)
	if err != nil {
		return err
	}

	if !resp.Ok {
		return fmt.Errorf(resp.Err)
	}

	return nil
}

// EventLoop ... Event loop for alert manager subsystem
func (am *alertManager) EventLoop(ctx context.Context) error {
	logger := logging.WithContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil

		case alert := <-am.alertTransit:
			logger.Info("received alert",
				zap.String(core.SUUIDKey, alert.SUUID.String()))

			alertDest, err := am.invariantAlertStore.GetAlertDestination(alert.SUUID)
			if err != nil {
				logger.Error("Could not determine alerting destination", zap.Error(err))
			}

			switch alertDest {
			case core.Slack: // TODO: add more alert destinations
				logger.Debug("Attempting to post alert to slack")

				err := am.handleSlackPost(alert)
				if err != nil {
					logger.Error("Could not post alert to slack", zap.Error(err))
				}
			}

			if alertDest == core.Slack {

			}
		}
	}

}
