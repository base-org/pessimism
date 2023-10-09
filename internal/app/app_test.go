package app_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
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
			RoutingCfgPath:          "",
			PagerdutyAlertEventsURL: "test",
			RoutingParams:           &core.AlertRoutingParams{},
		},
	}

	app, shutDown, err := app.NewPessimismApp(ctx, cfg)

	assert.NoError(t, err)

	err = app.Start()
	assert.NoError(t, err)

	shutDown()
}
