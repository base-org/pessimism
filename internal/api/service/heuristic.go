package service

import (
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
)

// ProcessHeuristicRequest ... Processes a heuristic request type
func (svc *PessimismService) ProcessHeuristicRequest(ir *models.SessionRequestBody) (core.SUUID, error) {
	if ir.MethodType() == models.Run { // Deploy heuristic session
		return svc.RunHeuristicSession(&ir.Params)
	}
	// TODO - Add support for other method types (ie. delete. update)

	return core.NilSUUID(), nil
}

// RunHeuristicSession ... Runs a heuristic session provided
func (svc *PessimismService) RunHeuristicSession(params *models.SessionRequestParams) (core.SUUID, error) {
	pConfig, err := svc.m.BuildPipelineCfg(params)
	if err != nil {
		return core.NilSUUID(), err
	}

	sConfig := params.SessionConfig()

	deployCfg, err := svc.m.BuildDeployCfg(pConfig, sConfig)
	if err != nil {
		return core.NilSUUID(), err
	}

	sUUID, err := svc.m.RunSession(deployCfg)
	if err != nil {
		return core.NilSUUID(), err
	}

	return sUUID, nil
}
