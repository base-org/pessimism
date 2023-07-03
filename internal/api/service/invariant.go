package service

import (
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
)

// ProcessInvariantRequest ... Processes an invariant request type
func (svc *PessimismService) ProcessInvariantRequest(ir *models.InvRequestBody) (core.SUUID, error) {
	if ir.MethodType() == models.Run { // Deploy invariant session
		return svc.RunInvariantSession(&ir.Params)
	}
	// TODO - Add support for other method types (ie. delete. update)

	return core.NilSUUID(), nil
}

// runInvariantSession ... Runs an invariant session provided
func (svc *PessimismService) RunInvariantSession(params *models.InvRequestParams) (core.SUUID, error) {

	pConfig, err := svc.m.BuildPipelineCfg(params)
	if err != nil {
		return core.NilSUUID(), err
	}

	sConfig := params.SessionConfig()

	sUUID, err := svc.m.RunInvSession(pConfig, sConfig)
	if err != nil {
		return core.NilSUUID(), err
	}

	return sUUID, nil
}
