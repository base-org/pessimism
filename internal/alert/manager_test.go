package alert_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestEventLoop(t *testing.T) {

	cfg := &config.Config{
		AlertConfig: &alert.Config{
			RoutingCfgPath:          "test_data/alert-routing-test.yaml",
			PagerdutyAlertEventsURL: "test",
			SNSConfig: &client.SNSConfig{
				TopicArn: "test",
			},
		},
	}

	ctx := context.Background()

	c := gomock.NewController(t)

	tests := []struct {
		name        string
		description string
		test        func(t *testing.T)
	}{
		{
			name:        "Test Low sev",
			description: "Test low sev alert sends to slack",
			test: func(t *testing.T) {
				cm := alert.NewRoutingDirectory(cfg.AlertConfig)
				sns := mocks.NewMockSNSClient(c)
				am := alert.NewManager(ctx, cfg.AlertConfig, cm)

				go func() {
					_ = am.EventLoop()
				}()

				defer func() {
					_ = am.Shutdown()
				}()

				ingress := am.Transit()

				cm.SetSNSClient(sns)
				cm.SetSlackClients([]client.SlackClient{mocks.NewMockSlackClient(c)}, core.LOW)

				alert := core.Alert{
					Sev:         core.LOW,
					HeuristicID: core.UUID{},
				}
				policy := &core.AlertPolicy{
					Sev: core.LOW.String(),
					Msg: "test",
				}

				err := am.AddSession(core.UUID{}, policy)
				assert.Nil(t, err)

				for _, cli := range cm.GetSlackClients(core.LOW) {
					sc, ok := cli.(*mocks.MockSlackClient)
					assert.True(t, ok)

					sc.EXPECT().PostEvent(gomock.Any(), gomock.Any()).Return(
						&client.AlertAPIResponse{
							Message: "test",
							Status:  core.SuccessStatus,
						}, nil).Times(1)
				}

				sns.EXPECT().PostEvent(gomock.Any(), gomock.Any()).Return(
					&client.AlertAPIResponse{
						Message: "test",
						Status:  core.SuccessStatus,
					}, nil).AnyTimes()

				sns.EXPECT().GetName().AnyTimes()

				ingress <- alert
				time.Sleep(1 * time.Second)
				id := core.NewUUID()
				alert = core.Alert{
					Sev:         core.UNKNOWN,
					HeuristicID: id,
				}
				ingress <- alert
				time.Sleep(1 * time.Second)

			},
		},
		{
			name:        "Test Medium sev",
			description: "Test medium sev alert sends to just PagerDuty",
			test: func(t *testing.T) {
				cm := alert.NewRoutingDirectory(cfg.AlertConfig)
				sns := mocks.NewMockSNSClient(c)
				am := alert.NewManager(ctx, cfg.AlertConfig, cm)

				go func() {
					_ = am.EventLoop()
				}()

				defer func() {
					_ = am.Shutdown()
				}()

				ingress := am.Transit()

				cm.SetPagerDutyClients([]client.PagerDutyClient{mocks.NewMockPagerDutyClient(c)}, core.MEDIUM)
				cm.SetSNSClient(sns)

				alert := core.Alert{
					Sev:         core.MEDIUM,
					HeuristicID: core.UUID{},
				}
				policy := &core.AlertPolicy{
					Sev: core.MEDIUM.String(),
					Msg: "test",
				}

				err := am.AddSession(core.UUID{}, policy)
				assert.Nil(t, err)

				for _, cli := range cm.GetPagerDutyClients(core.MEDIUM) {
					pdc, ok := cli.(*mocks.MockPagerDutyClient)
					assert.True(t, ok)

					pdc.EXPECT().PostEvent(gomock.Any(), gomock.Any()).Return(
						&client.AlertAPIResponse{
							Message: "test",
							Status:  core.SuccessStatus,
						}, nil).Times(1)
				}

				sns.EXPECT().PostEvent(gomock.Any(), gomock.Any()).Return(
					&client.AlertAPIResponse{
						Message: "test",
						Status:  core.SuccessStatus,
					}, nil).AnyTimes()

				sns.EXPECT().GetName().AnyTimes()

				ingress <- alert
				time.Sleep(1 * time.Second)
				id := core.UUID{}
				alert = core.Alert{
					Sev:         core.UNKNOWN,
					HeuristicID: id,
				}
				ingress <- alert
				time.Sleep(1 * time.Second)

			},
		},
		{
			name:        "Test High sev",
			description: "Test high sev alert sends to both slack and PagerDuty",
			test: func(t *testing.T) {
				cm := alert.NewRoutingDirectory(cfg.AlertConfig)
				sns := mocks.NewMockSNSClient(c)
				am := alert.NewManager(ctx, cfg.AlertConfig, cm)

				go func() {
					_ = am.EventLoop()
				}()

				defer func() {
					_ = am.Shutdown()
				}()

				ingress := am.Transit()

				cm.SetSlackClients([]client.SlackClient{mocks.NewMockSlackClient(c), mocks.NewMockSlackClient(c)}, core.HIGH)
				cm.SetPagerDutyClients([]client.PagerDutyClient{mocks.NewMockPagerDutyClient(c), mocks.NewMockPagerDutyClient(c)}, core.HIGH)
				cm.SetSNSClient(sns)

				alert := core.Alert{
					Sev:         core.HIGH,
					HeuristicID: core.UUID{},
				}
				policy := &core.AlertPolicy{
					Sev: core.HIGH.String(),
					Msg: "test",
				}
				err := am.AddSession(core.UUID{}, policy)
				assert.Nil(t, err)

				for _, cli := range cm.GetPagerDutyClients(core.HIGH) {
					pdc, ok := cli.(*mocks.MockPagerDutyClient)
					assert.True(t, ok)

					pdc.EXPECT().PostEvent(gomock.Any(), gomock.Any()).Return(
						&client.AlertAPIResponse{
							Message: "test",
							Status:  core.SuccessStatus,
						}, nil)
				}

				for _, cli := range cm.GetSlackClients(core.HIGH) {
					sc, ok := cli.(*mocks.MockSlackClient)
					assert.True(t, ok)
					sc.EXPECT().PostEvent(gomock.Any(), gomock.Any()).Return(
						&client.AlertAPIResponse{
							Message: "test",
							Status:  core.SuccessStatus,
						}, nil)
				}

				sns.EXPECT().PostEvent(gomock.Any(), gomock.Any()).Return(
					&client.AlertAPIResponse{
						Message: "test",
						Status:  core.SuccessStatus,
					}, nil).AnyTimes()

				sns.EXPECT().GetName().AnyTimes()

				ingress <- alert
				time.Sleep(1 * time.Second)
				id := core.UUID{}
				alert = core.Alert{
					Sev:         core.UNKNOWN,
					HeuristicID: id,
				}
				ingress <- alert
				time.Sleep(1 * time.Second)
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%s:%d", test.name, i), func(t *testing.T) {
			test.test(t)
		})
	}

}
