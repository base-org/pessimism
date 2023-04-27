package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"time"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

const (
	cfgPath = "config.env"
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

func initializeAndRunServer(ctx context.Context, cfgPath config.FilePath) (*server.Server, func(), error) {
	cfg := config.NewConfig(cfgPath)

	apiService := service.New()
	handler, err := handlers.New(apiService)
	if err != nil {
		return nil, nil, err
	}
	server, cleanup, err := server.New(ctx, cfg.ServerConfig, handler)
	if err != nil {
		return nil, nil, err
	}
	return server, func() {
		cleanup()
	}, nil
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
	defer cancel()
	cfg := config.NewConfig(cfgPath)

	wg := &sync.WaitGroup{}

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logger := logging.WithContext(appCtx)

	logger.Info("pessimism boot up")

	etlManager, shutDownETL := pipeline.NewManager(appCtx)
	outChan, err := setupExampleETL(cfg, etlManager)
	if err != nil {
		panic(err)
	}

	wg.Add(1)

	go func() {
		etlManager.EventLoop(appCtx)
		wg.Done()
	}()

	go func() {
		server, shutDownServer, err := initializeAndRunServer(appCtx, cfgPath)

		if err != nil {
			logger.Error("Error obtained trying to start server", zap.Error(err))
			panic(err)
		}

		server.Stop(func() {
			shutDownETL()
			shutDownServer()
		})
	}()

	processExampleETL(appCtx, outChan)
}
