package service

import (
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
)

// ProcessInvariantRequest ... Processes an invariant request type
func (svc *PessimismService) ProcessInvariantRequest(ir models.InvRequestBody) (core.InvSessionUUID, error) {
	if ir.MethodType() == models.Run { // Deploy invariant session
		return svc.runInvariantSession(ir.Params)
	}
	// TODO - Add support for other method types (ie. delete. update)

	return core.NilInvariantUUID(), nil
}

// runInvariantSession ... Runs an invariant session provided
func (svc *PessimismService) runInvariantSession(params models.InvRequestParams) (core.InvSessionUUID, error) {
	inv, err := registry.GetInvariant(params.InvariantType(), params.SessionParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	// TODO(#53): API Request Validation Submodule
	endpoint, err := svc.cfg.GetEndpointForNetwork(params.NetworkType())
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	pollInterval, err := svc.cfg.GetPollIntervalForNetwork(params.NetworkType())
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	pConfig := params.GeneratePipelineConfig(endpoint, pollInterval, inv.InputType())
	sConfig := params.SessionConfig()

	sUUID, err := svc.m.StartInvSession(pConfig, sConfig)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	return sUUID, nil
}
