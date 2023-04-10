package main

import (
	"context"
	"log"
	"time"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	/*
		This a simple experimental POC showcasing a pipeline DAG with two register pipelines:
							->
		(Geth Block Node) -^

		This is done to:
		A) Prove that the Oracle and Pipe components operate as expected and are able to channel data between each other
		B) Reason about component construction to better understand how to automate register pipeline creation
		C) Demonstrate a lightweight MVP for the system
	*/

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig("config.env")
	l1OracleCfg := &config.OracleConfig{
		RPCEndpoint: cfg.L1RpcEndpoint,
		StartHeight: nil,
		EndHeight:   nil}

	registerCfg1 := &config.PipelineConfig{
		DataType:     core.ContractCreateTX,
		PipelineType: core.Live,
		OracleCfg:    l1OracleCfg,
	}

	registerCfg2 := &config.PipelineConfig{
		DataType:     core.BlackholeTX,
		PipelineType: core.Live,
		OracleCfg:    l1OracleCfg,
	}

	etlManager := pipeline.NewManager(appCtx)

	pID, err := etlManager.CreateRegisterPipeline(appCtx, registerCfg1)
	if err != nil {
		panic(err)
	}

	pID2, err := etlManager.CreateRegisterPipeline(appCtx, registerCfg2)
	if err != nil {
		panic(err)
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

		switch td.Type {

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
			break

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
			break
		}
	}
}
