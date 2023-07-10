package config_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	// Ensure that root level template file can be successfully parsed into a config struct
	cfg := config.NewConfig("../../config.env.template")
	assert.NotNil(t, cfg, "Config should not be nil")
}
