package service

import (
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// ProcessInvariantRequest ... Processes an invariant request type
func (svc *PessimismService) ProcessInvariantRequest(ir models.InvRequestBody) (core.InvSessionUUID, error) {
	if ir.Method == models.Run { // Deploy invariant session
		return svc.runInvariantSession(ir.Params)
	}
	// TODO - Add support for other run types

	return core.NilInvariantUUID(), nil
}

// runInvariantSession ... Runs an invariant session provided
func (svc *PessimismService) runInvariantSession(params models.InvRequestParams) (core.InvSessionUUID, error) {
	logger := logging.WithContext(svc.ctx)

	inv, err := registry.GetInvariant(params.InvType, params.SessionParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	endpoint, err := svc.cfg.GetEndpointForNetwork(params.Network)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	pollInterval, err := svc.cfg.GetPollIntervalForNetwork(params.Network)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	pConfig := params.GeneratePipelineConfig(endpoint, pollInterval, inv.InputType())

	pID, err := svc.etlManager.CreateDataPipeline(pConfig)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	logger.Info("Created etl pipeline",
		zap.String(core.PUUIDKey, pID.String()))

	invID, err := svc.engineManager.DeployInvariantSession(params.Network, pID, params.InvType,
		params.PType, params.SessionParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}
	logger.Info("Deployed invariant session", zap.String(core.SUUIDKey, invID.String()))

	if err = svc.etlManager.RunPipeline(pID); err != nil {
		return core.NilInvariantUUID(), err
	}

	return invID, nil
}
