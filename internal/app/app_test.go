package app_test

import (
	"context"
	"github.com/base-org/pessimism/internal/client/alert_clients"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/metrics"

	"github.com/stretchr/testify/assert"
)

func Test_AppFlow(t *testing.T) {

	ctx := context.Background()

	cfg := &config.Config{
		ServerConfig: &server.Config{
			Host: "localhost",
			Port: 8080,
		},

		MetricsConfig: &metrics.Config{
			Enabled: false,
		},
		AlertConfig: &alert.Config{
			SlackConfig:        &alert_clients.SlackConfig{},
			MediumPagerDutyCfg: &alert_clients.PagerDutyConfig{},
			HighPagerDutyCfg:   &alert_clients.PagerDutyConfig{},
		},
	}

	app, shutDown, err := app.NewPessimismApp(ctx, cfg)

	assert.NoError(t, err)

	err = app.Start()
	assert.NoError(t, err)

	shutDown()
}
