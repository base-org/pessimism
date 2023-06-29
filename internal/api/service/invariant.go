package service

import (
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
)

// ProcessInvariantRequest ... Processes an invariant request type
func (svc *PessimismService) ProcessInvariantRequest(ir models.InvRequestBody) (core.SUUID, error) {
	if ir.MethodType() == models.Run { // Deploy invariant session
		return svc.RunInvariantSession(ir.Params)
	}
	// TODO - Add support for other method types (ie. delete. update)

	return core.NilSUUID(), nil
}

// runInvariantSession ... Runs an invariant session provided
func (svc *PessimismService) RunInvariantSession(params models.InvRequestParams) (core.SUUID, error) {
	inv, err := registry.GetInvariant(svc.ctx, params.InvariantType(), params.SessionParams)
	if err != nil {
		return core.NilSUUID(), err
	}

	pollInterval, err := svc.cfg.GetPollIntervalForNetwork(params.NetworkType())
	if err != nil {
		return core.NilSUUID(), err
	}

	pConfig := params.GeneratePipelineConfig(pollInterval, inv.InputType())
	sConfig := params.SessionConfig()

	sUUID, err := svc.m.StartInvSession(pConfig, sConfig)
	if err != nil {
		return core.NilSUUID(), err
	}

	return sUUID, nil
}
