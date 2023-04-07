package main

import (
	"context"
	"log"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/models"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	/*
		This a simple experimental POC showcasing an implicit CONTRACT_CREATE_TX register pipeline

		This is done to:
		A) Prove that the Oracle and Pipe components operate as expected and are able to channel data between each other
		B) Reason about component construction to better understand how to automate register pipeline creation
		C) Demonstrate a lightweight MVP for the system
	*/

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig("../../config.env")
	l1OracleCfg := &config.OracleConfig{
		RPCEndpoint: cfg.L1RpcEndpoint,
		StartHeight: nil,
		EndHeight:   nil}

	registerCfg := &config.RegisterPipelineConfig{
		DataType:     registry.ContractCreateTX,
		PipelineType: models.Live,
		OracleCfg:    l1OracleCfg,
	}

	pipeLineDAG := pipeline.NewDAG()

	pID, err := pipeLineDAG.CreateRegisterPipeline(appCtx, registerCfg)
	if err != nil {
		panic(err)
	}

	err = pipeLineDAG.RunPipeline(pID)
	if err != nil {
		panic(err)
	}

	outChan := models.NewTransitChannel()

	if err := pipeLineDAG.AddPipelineDirective(pID, models.NilID(), outChan); err != nil {
		panic(err)
	}

	log.Printf("===============================================")
	log.Printf("Reading layer 1 EVM blockchain for live contract creation txs")
	log.Printf("===============================================")

	for td := range outChan {
		log.Printf("===============================================")
		log.Printf("Received Contract Creation (CREATE) Transaction %+v", td)
		log.Printf("===============================================")

		parsedTx, success := td.Value.(*types.Transaction)
		if !success {
			log.Printf("Could not parse transaction value")
		} else {
			log.Printf("As parsed transaction %+v", parsedTx)
		}
	}
}
