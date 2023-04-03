package registry

import (
	"context"
	"fmt"
	"strconv"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/models"
)

func ToOracleType(pt models.PipelineType) (pipeline.OracleType, error) {
	switch pt {
	case models.Backtest:
		return pipeline.BacktestOracle, nil

	case models.LiveTest:
		return pipeline.LiveOracle, nil

	case models.MockTest:
		return pipeline.UnknownOracle, fmt.Errorf("Conveyor component has yet to be implemented")

	default:
		return pipeline.UnknownOracle, fmt.Errorf("Unsupported type provided")

	}
}

type RegisterPipeline struct {
	ctx        context.Context
	components pipeline.Component
	dataType   models.RegisterType
	pipeChans  []models.TransitChannel
}

func NewRegisterPipeline(ctx context.Context, cfg *config.RegisterPipelineConfig) (*RegisterPipeline, error) {
	register, err := GetRegister(cfg.DataType)
	if err != nil {
		return nil, err
	}

	rp := &RegisterPipeline{pipeChans: make([]chan models.TransitData, 0)}

	lastComponent, err := rp.inferComponent(ctx, cfg, register)
	if err != nil {
		return nil, err
	}

	components := make(pipeline.Component, 0)

	compID, err := strconv.Atoi(string(register.DataType))
	if err != nil {
		return nil, err
	}

	components[compID] = lastComponent

	for _, depRegister := range register.Dependencies {

		depComponent, err := rp.inferComponent(ctx, cfg, depRegister)
		if err != nil {
			return nil, err
		}
		compID, err := strconv.Atoi(string(register.DataType))
		if err != nil {
			return nil, err
		}

		components[compID] = depComponent
	}

	return &RegisterPipeline{
		ctx:        ctx,
		components: components,
		dataType:   cfg.DataType,
	}, nil
}

func (rp *RegisterPipeline) RunPipeline() {
	pipeIndex := 0

	for compID, component := range rp.components {
		if component.Type() == models.Pipe {
			ch := rp.pipeChans[pipeIndex]
			compID := append
		}

	}

}

func (rp *RegisterPipeline) inferComponent(ctx context.Context, cfg *config.RegisterPipelineConfig, register *DataRegister) (pipeline.Component, error) {

	switch register.ComponentType {
	case models.Oracle:
		init, success := register.ComponentConstructor.(pipeline.OracleConstructor)
		if !success {
			return nil, fmt.Errorf("Could not cast constructor to oracle constructor type")
		}

		ot, err := ToOracleType(cfg.PipelineType)
		if err != nil {
			return nil, err
		}

		// NOTE ... We assume at most 1 oracle per register pipeline
		return init(ctx, ot, cfg.OracleCfg)

	case models.Pipe:
		init, success := register.ComponentConstructor.(pipeline.PipeConstructorFunc)
		if !success {
			return nil, fmt.Errorf("Could not cast constructor to pipe constructor type")
		}
		ch := models.NewTransitChannel()
		rp.pipeChans = append(rp.pipeChans, ch)
		return init(ctx, ch)

	case models.Conveyor:
		return nil, fmt.Errorf("Conveyor component has yet to be implemented")

	default:
		return nil, fmt.Errorf("Unknown Component type provided")
	}
}
