package service

import (
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// ProcessInvariantRequest ... Processes an invariant request type
func (svc *PessimismService) ProcessInvariantRequest(ir models.InvRequestBody) (core.SUUID, error) {
	if ir.MethodType() == models.Run { // Deploy invariant session
		return svc.RunInvariantSession(ir.Params)
	}
	// TODO - Add support for other method types (ie. delete. update)

	return core.NilSUUID(), nil
}

// runInvariantSession ... Runs an invariant session
func (svc *PessimismService) RunInvariantSession(params models.InvRequestParams) (core.SUUID, error) {
	logger := logging.WithContext(svc.ctx)

	inv, err := registry.GetInvariant(params.InvariantType(), params.SessionParams)
	if err != nil {
		return core.NilSUUID(), err
	}

	// TODO(#53): API Request Validation Submodule
	endpoint, err := svc.cfg.GetEndpointForNetwork(params.NetworkType())
	if err != nil {
		return core.NilSUUID(), err
	}

	pollInterval, err := svc.cfg.GetPollIntervalForNetwork(params.NetworkType())
	if err != nil {
		return core.NilSUUID(), err
	}

	pConfig := params.GeneratePipelineConfig(endpoint, pollInterval, inv.InputType())

	pUUID, err := svc.etlManager.CreateDataPipeline(pConfig)
	if err != nil {
		return core.NilSUUID(), err
	}

	inRegister, err := svc.etlManager.GetRegister(pConfig.DataType)
	if err != nil {
		return core.NilSUUID(), err
	}

	logger.Info("Created etl pipeline",
		zap.String(core.PUUIDKey, pUUID.String()))

	invID, err := svc.engineManager.DeployInvariantSession(params.NetworkType(), pUUID, params.InvariantType(),
		params.PiplineType(), params.SessionParams, inRegister)
	if err != nil {
		return core.NilSUUID(), err
	}
	logger.Info("Deployed invariant session", zap.String(core.SUUIDKey, invID.String()))

	err = svc.alertManager.AddInvariantSession(invID, params.AlertingDestType())
	if err != nil {
		return core.NilSUUID(), err
	}

	if err = svc.etlManager.RunPipeline(pUUID); err != nil {
		return core.NilSUUID(), err
	}

	return invID, nil
}
