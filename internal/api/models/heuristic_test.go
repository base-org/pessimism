package models_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_SessionRequestParams(t *testing.T) {

	// Use a single instance of SessionRequestParams for tests
	irp := &models.SessionRequestParams{
		SessionParams: map[string]interface{}{
			"test": "test",
		},
		Network:       core.Layer1.String(),
		PType:         core.Live.String(),
		HeuristicType: core.BalanceEnforcement.String(),
	}

	// Ensure that the heuristic request params are set correctly
	params := irp.Params()
	v, err := params.Value("test")
	assert.NoError(t, err)
	assert.Equal(t, v, "test")

	// Ensure that network type is set correctly
	n := irp.NetworkType()
	assert.Equal(t, n, core.Layer1)

	// Ensure that pipeline type is set correctly
	pt := irp.PathType()
	assert.Equal(t, pt, core.Live)

	// Ensure that heuristic type is set correctly
	it := irp.Heuristic()
	assert.Equal(t, it, core.BalanceEnforcement)

	// Ensure that the pipeline config is set correctly
	pConfig := irp.GeneratePipelineConfig(0, 0)
	assert.Equal(t, pConfig.Network, core.Layer1)
	assert.Equal(t, pConfig.PathType, core.Live)

	sConfig := irp.SessionConfig()
	assert.Equal(t, sConfig.Type, core.BalanceEnforcement)
	assert.Equal(t, sConfig.PT, core.Live)
	assert.Equal(t, sConfig.Params, params)
}

func Test_HeuristicRequestBody(t *testing.T) {
	irb := &models.SessionRequestBody{
		Method: "test",
		Params: models.SessionRequestParams{},
	}

	// Ensure clone works
	clone := irb.Clone()
	assert.Equal(t, clone.Method, irb.Method)
	assert.Equal(t, clone.Params, irb.Params)

	// Ensure that method type works
	assert.Equal(t, irb.MethodType(), models.HeuristicMethod(0))
}
