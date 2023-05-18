package alert

import (
	"context"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// AlertingManager ...
type AlertingManager interface {
	AddInvariantSession(core.InvSessionUUID, core.AlertDestination) error
	EventLoop(ctx context.Context) error
	Transit() chan core.Alert
}

// alertManager ...
type alertManager struct {
	sc                  client.SlackClient
	invariantAlertStore AlertStore

	alertTransit chan core.Alert
}

// NewManager ... Instantiates a new alert manager
func NewManager(sc client.SlackClient) (AlertingManager, func()) {
	am := &alertManager{
		sc:                  sc,
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

// Transit ... Returns inter-subsystem transit channel
func (am *alertManager) Transit() chan core.Alert {
	return am.alertTransit
}

// EventLoop ... Event loop for alert manager
func (am *alertManager) EventLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case alert := <-am.alertTransit:
			alertDest, err := am.invariantAlertStore.GetAlertDestination(alert.SUUID)
			if err != nil {
				return err
			}

			if alertDest == core.Slack { // TODO - response validation
				_, err := am.sc.PostAlert(alert.Content)
				if err != nil {
					logging.NoContext().Error("failed to post alert to slack", zap.Error(err))
				}
			}
		}
	}

}
