//go:generate mockgen -package mocks --destination ../mocks/alert_manager.go --mock_names Manager=AlertManager . Manager

package alert

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Manager ... Interface for alert manager
type Manager interface {
	AddInvariantSession(core.SUUID, core.AlertDestination) error
	Transit() chan core.Alert

	core.Subsystem
}

// alertManager ... Alert manager implementation
type alertManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	sc           client.SlackClient
	store        Store
	interpolator Interpolator

	alertTransit chan core.Alert
}

// NewManager ... Instantiates a new alert manager
func NewManager(ctx context.Context, sc client.SlackClient) Manager {
	// NOTE - Consider constructing dependencies in higher level
	// abstraction and passing them in

	ctx, cancel := context.WithCancel(ctx)

	am := &alertManager{
		ctx:    ctx,
		cancel: cancel,

		sc:           sc,
		interpolator: NewInterpolator(),
		store:        NewStore(),
		alertTransit: make(chan core.Alert),
	}

	return am
}

// AddInvariantSession ... Adds an invariant session to the alert manager store
func (am *alertManager) AddInvariantSession(sUUID core.SUUID, alertDestination core.AlertDestination) error {
	return am.store.AddAlertDestination(sUUID, alertDestination)
}

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

// EventLoop ... Event loop for alert manager subsystem
func (am *alertManager) EventLoop() error {
	logger := logging.WithContext(am.ctx)

	for {
		select {
		case <-am.ctx.Done():
			return nil

		case alert := <-am.alertTransit:
			logger.Info("received alert",
				zap.String(core.SUUIDKey, alert.SUUID.String()))

			alertDest, err := am.store.GetAlertDestination(alert.SUUID)
			if err != nil {
				logger.Error("Could not determine alerting destination", zap.Error(err))
				continue
			}

			switch alertDest {
			case core.Slack: // TODO: add more alert destinations
				logger.Debug("Attempting to post alert to slack")

				err := am.handleSlackPost(alert)
				if err != nil {
					logger.Error("Could not post alert to slack", zap.Error(err))
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
