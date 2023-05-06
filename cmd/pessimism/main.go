package main

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/api/service"
	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	inv_registry "github.com/base-org/pessimism/internal/engine/registry"

	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/logging"
)

const (
	cfgPath = "config.env"
)

func setupExampleETL(cfg *config.Config, m *pipeline.Manager) (chan core.TransitData, error) {
	l1OracleCfg := &config.OracleConfig{
		RPCEndpoint: cfg.L1RpcEndpoint,
		StartHeight: big.NewInt(17170900),
		EndHeight:   nil}

	pipelineCfg1 := &config.PipelineConfig{
		Network:      core.Layer1,
		DataType:     core.ContractCreateTX,
		PipelineType: core.Live,
		OracleCfg:    l1OracleCfg,
	}

	pID, err := m.CreateDataPipeline(pipelineCfg1)
	if err != nil {
		return nil, err
	}

	outChan := core.NewTransitChannel()

	if err := m.AddPipelineDirective(pID, core.NilComponentUUID(), outChan); err != nil {
		return nil, err
	}

	if err := m.RunPipeline(pID); err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 1)

	return outChan, nil
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
	wg := &sync.WaitGroup{}

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewConfig(cfgPath)

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
		defer wg.Done()

		etlManager.EventLoop(appCtx)

	}()

	invStore := engine.NewInvariantStore()
	err = invStore.AddInvariant(core.ContractCreateTX, inv_registry.NewExampleInvariant(
		&inv_registry.ExampleInvConfig{
			FromAddress: "0x1DcBbDc86c0fb66788323b35Fe9046C6c761E418",
		},
	))

	if err != nil {
		panic(err)
	}

	riskEngine := engine.NewEngine(outChan, invStore)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := riskEngine.EventLoop(appCtx); err != nil {
			panic(err)
		}
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

	wg.Wait()
}
