package models_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_InvRequestParams(t *testing.T) {

	// Use a single instance of InvRequestParams for tests
	irp := &models.InvRequestParams{
		SessionParams: map[string]interface{}{
			"test": "test",
		},
		Network: core.Layer1.String(),
		PType:   core.Live.String(),
		InvType: core.BalanceEnforcement.String(),
	}

	// Ensure that the invariant request params are set correctly
	params := irp.Params()
	v, err := params.Value("test")
	assert.NoError(t, err)
	assert.Equal(t, v, "test")

	// Ensure that network type is set correctly
	n := irp.NetworkType()
	assert.Equal(t, n, core.Layer1)

	// Ensure that pipeline type is set correctly
	pt := irp.PipelineType()
	assert.Equal(t, pt, core.Live)

	// Ensure that invariant type is set correctly
	it := irp.InvariantType()
	assert.Equal(t, it, core.BalanceEnforcement)

	// Ensure that the pipeline config is set correctly
	pConfig := irp.GeneratePipelineConfig(0, 0)
	assert.Equal(t, pConfig.Network, core.Layer1)
	assert.Equal(t, pConfig.PipelineType, core.Live)

	sConfig := irp.SessionConfig()
	assert.Equal(t, sConfig.Type, core.BalanceEnforcement)
	assert.Equal(t, sConfig.PT, core.Live)
	assert.Equal(t, sConfig.Params, params)
}
