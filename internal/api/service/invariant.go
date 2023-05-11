package service

import (
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

const (
	RetryCount = 3
)

func (svc *PessimismService) ProcessInvariantRequest(ir models.InvRequestBody) (core.InvSessionUUID, error) {
	if ir.Method == models.Run {
		return svc.runInvariantSession(&ir)
	}
	// TODO - Add support for other run types

	return core.NilInvariantUUID(), nil
}

func (svc *PessimismService) runInvariantSession(ir *models.InvRequestBody) (core.InvSessionUUID, error) {
	logger := logging.WithContext(svc.ctx)

	inv, err := registry.GetInvariant(ir.Params.InvType, ir.Params.SessionParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	endPoint, err := svc.cfg.GetEndpointFromNetwork(ir.Params.Network)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	oCfg := core.OracleConfig{
		RPCEndpoint:  endPoint,
		StartHeight:  ir.Params.StartHeight,
		EndHeight:    ir.Params.EndHeight,
		NumOfRetries: RetryCount, // TODO - Make configurable through env file
	}

	pCfg := &core.PipelineConfig{
		Network:      ir.Params.Network,
		DataType:     inv.InputType(),
		PipelineType: ir.Params.PType,
		OracleCfg:    &oCfg,
	}

	logger.Info("Creating data pipeline")
	pID, err := svc.etlManager.CreateDataPipeline(pCfg)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	invID, err := svc.engineManager.DeployInvariantSession(ir.Params.Network, pID, ir.Params.InvType,
		ir.Params.PType, ir.Params.SessionParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}
	logger.Info("Deployed invariant session", zap.String("invariant_uuid", invID.String()))

	if err = svc.etlManager.RunPipeline(pID); err != nil {
		return core.NilInvariantUUID(), err
	}

	return invID, nil
}
