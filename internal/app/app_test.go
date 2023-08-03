package app_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_AppFlow(t *testing.T) {

	ctx := mocks.Context(context.Background(), gomock.NewController(t))

	cfg := &config.Config{
		ServerConfig: &server.Config{
			Host: "localhost",
			Port: 8080,
		},

		MetricsConfig: &metrics.Config{
			Enabled: false,
		},
		AlertConfig: &alert.Config{
			SlackConfig:     &client.SlackConfig{},
			PagerdutyConfig: &client.PagerdutyConfig{},
		},
	}

	app, shutDown, err := app.NewPessimismApp(ctx, cfg)

	assert.NoError(t, err)

	err = app.Start()
	assert.NoError(t, err)

	shutDown()
}
