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

func (svc *PessimismService) ProcessInvariantRequest(ir models.InvRequestBody) (core.InvariantUUID, error) {
	if ir.Method == models.Run {
		return svc.runInvariant(&ir)
	}

	return core.NilInvariantUUID(), nil
}

func (svc *PessimismService) runInvariant(ir *models.InvRequestBody) (core.InvariantUUID, error) {
	logger := logging.WithContext(svc.ctx)

	inv, err := registry.GetInvariant(ir.Params.InvType, ir.Params.InvParams)
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

	invID, err := svc.engineManager.DeployInvariantSession(ir.Params.Network, ir.Params.InvType,
		ir.Params.PType, ir.Params.InvParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}
	logger.Info("Deployed invariant session", zap.String("invariant_id", invID.String()))

	logger.Info("Creating register pipeline")
	pID, err := svc.etlManager.CreateDataPipeline(pCfg)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	logger.Info("Creating register pipeline")

	if err = svc.etlManager.RunPipeline(pID); err != nil {
		return core.NilInvariantUUID(), err
	}

	return invID, nil
}
