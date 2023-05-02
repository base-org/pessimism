package main

import (
	"context"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

func setupExampleETL(cfg *config.Config, m *pipeline.Manager) (chan core.TransitData, error) {
	l1OracleCfg := &config.OracleConfig{
		RPCEndpoint: cfg.L1RpcEndpoint,
		StartHeight: nil,
		EndHeight:   nil}

	pipelineCfg1 := &config.PipelineConfig{
		Network:      core.Layer1,
		DataType:     core.ContractCreateTX,
		PipelineType: core.Live,
		OracleCfg:    l1OracleCfg,
	}

	pipelineCfg2 := &config.PipelineConfig{
		Network:      core.Layer1,
		DataType:     core.BlackholeTX,
		PipelineType: core.Live,
		OracleCfg:    l1OracleCfg,
	}

	pID, err := m.CreateDataPipeline(pipelineCfg1)
	if err != nil {
		return nil, err
	}

	pID2, err := m.CreateDataPipeline(pipelineCfg2)
	if err != nil {
		return nil, err
	}

	outChan := core.NewTransitChannel()

	if err := m.AddPipelineDirective(pID, core.NilComponentUUID(), outChan); err != nil {
		return nil, err
	}

	if err := m.AddPipelineDirective(pID2, core.NilComponentUUID(), outChan); err != nil {
		return nil, err
	}

	if err := m.RunPipeline(pID); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 1)

	if err := m.RunPipeline(pID2); err != nil {
		return nil, err
	}

	return outChan, nil
}

func processExampleETL(ctx context.Context, outChan chan core.TransitData) {
	logger := logging.WithContext(ctx)

	logger.Info("Reading layer 1 EVM blockchain for live contract creation txs")

	for td := range outChan {
		switch td.Type { //nolint:exhaustive // checks for all transit data types are unnecessary here
		case core.ContractCreateTX:

			parsedTx, success := td.Value.(*types.Transaction)
			if !success {
				logger.Error("Could not parse transaction value")
			} else {
				logger.Info("Received Contract Creation (CREATE) Transaction", zap.String("tx", fmt.Sprintf("%+v", parsedTx)))
			}

		case core.BlackholeTX:

			parsedTx, success := td.Value.(*types.Transaction)
			if !success {
				logger.Error("Could not parse transaction value")
			} else {
				logger.Info("Received Blackhole (NULL) Transaction", zap.String("tx", fmt.Sprintf("%+v", parsedTx)))
			}
		}
	}
}

// TODO(#34): No Documentation Exists Specifying How to Run & Test Service
func main() {
	/*
		This a simple experimental POC showcasing a pipeline DAG with two register pipelines that use overlapping components:
							    -> (C1)(Contract Create TX Pipe)
		(C0)(Geth Block Node) --
							    -> (C3)(Blackhole Address Tx Pipe)
		This is done to:
		A) Prove that the Oracle and Pipe components operate as expected and are able to channel data between each other
		B) Showcase a minimal example of the Pipeline DAG that can leverage overlapping register components to avoid
			duplication when necessary
		C) Demonstrate a lightweight MVP for the system
	*/

	appCtx, cancel := context.WithCancel(context.Background())

	localState := state.NewMemState()
	appCtx = context.WithValue(appCtx, "state", localState)

	defer cancel()
	cfg := config.NewConfig("config.env")

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logger := logging.WithContext(appCtx)

	logger.Info("pessimism boot up")

	etlManager := pipeline.NewManager(appCtx)
	outChan, err := setupExampleETL(cfg, etlManager)
	if err != nil {
		panic(err)
	}

	go func() {
		etlManager.EventLoop(appCtx)
	}()

	processExampleETL(appCtx, outChan)
}
