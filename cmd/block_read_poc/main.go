package main

import (
	"context"
	"log"
	"time"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	/*
		This a simple experimental POC showcasing a pipeline DAG with two register pipelines that use overlapping components:
							-> (C1)(Contract Create TX Pipe)
		(C0)(Geth Block Node) --
							-> (C3)(Blackhole Address Tx Pipe)
		This is done to:
		A) Prove that the Oracle and Pipe components operate as expected and are able to channel data between each other
		B) Showcase a minimal example of the Pipeline DAG that can leverage overlapping register components to avoid duplication
		   when necessary
		C) Demonstrate a lightweight MVP for the system
	*/

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig("config.env")

	logging.NewLogger(cfg.LoggerConfig, cfg.IsProduction())

	logging.NoContext().Info("pessimism boot up")

	l1OracleCfg := &config.OracleConfig{
		RPCEndpoint: cfg.L1RpcEndpoint,
		StartHeight: nil,
		EndHeight:   nil}

	pipelineCfg1 := &config.PipelineConfig{
		DataType:     core.ContractCreateTX,
		PipelineType: core.Live,
		OracleCfg:    l1OracleCfg,
	}

	pipelineCfg2 := &config.PipelineConfig{
		DataType:  core.BlackholeTX,
		OracleCfg: l1OracleCfg,
	}

	outChan := core.NewTransitChannel()

	if err := etlManager.AddPipelineDirective(pID, core.NilCompID(), outChan); err != nil {
		panic(err)
	}

	if err := etlManager.AddPipelineDirective(pID2, core.NilCompID(), outChan); err != nil {
		panic(err)
	}

	err = etlManager.RunPipeline(pID)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 1)

	err = etlManager.RunPipeline(pID2)
	if err != nil {
		panic(err)
	}

	log.Printf("===============================================")
	log.Printf("Reading layer 1 EVM blockchain for live contract creation txs")
	log.Printf("===============================================")

	for td := range outChan {
		switch td.Type { //nolint:exhaustive // checks for all transit data types are unnecessary here
		case core.ContractCreateTX:
			log.Printf("===============================================")
			log.Printf("Received Contract Creation (CREATE) Transaction %+v", td)
			log.Printf("===============================================")

			parsedTx, success := td.Value.(*types.Transaction)
			if !success {
				log.Printf("Could not parse transaction value")
			} else {
				log.Printf("As parsed transaction %+v", parsedTx)
			}

		case core.BlackholeTX:
			log.Printf("===============================================")
			log.Printf("Received Blackhole (NULL) Transaction %+v", td)
			log.Printf("===============================================")

			parsedTx, success := td.Value.(*types.Transaction)
			if !success {
				log.Printf("Could not parse transaction value")
			} else {
				log.Printf("As parsed transaction %+v", parsedTx)
			}
		}
	}
}
