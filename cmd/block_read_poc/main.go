package main

import (
	"context"
	"log"
	"sync"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/conduit/pipeline"
	"github.com/base-org/pessimism/internal/conduit/registry"
	"github.com/base-org/pessimism/internal/config"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	outChanID   = 0x420
	interChanID = 0x42
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

	cfg := config.NewConfig("config.env")
	l1OracleCfg := &config.OracleConfig{
		RPCEndpoint: cfg.L1RpcEndpoint,
		StartHeight: nil,
		EndHeight:   nil}

	// 1. Configure blackhole tx pipe component
	createRegister, err := registry.GetRegister(registry.ContractCreateTX)
	if err != nil {
		panic(err)
	}

	initPipe, success := createRegister.ComponentConstructor.(pipeline.PipeConstructorFunc)
	if !success {
		panic("Could not read component constructor Pipe constructor type")
	}

	inputChan := make(chan models.TransitData)

	createTxPipe, err := initPipe(appCtx, inputChan)
	if err != nil {
		panic(err)
	}

	register, err := registry.GetRegister(registry.GethBlock)
	if err != nil {
		panic(err)
	}

	init, success := register.ComponentConstructor.(pipeline.OracleConstructor)
	if !success {
		panic("Could not read constructor value")
	}

	go func() {
		if routineErr := createTxPipe.EventLoop(); routineErr != nil {
			log.Printf("Error received from oracle event loop %e", err)
		}
	}()

	l1Oracle, err := init(appCtx, pipeline.LiveOracle, l1OracleCfg)
	if err != nil {
		panic(err)
	}

	if err := l1Oracle.AddDirective(interChanID, inputChan); err != nil {
		panic(err)
	}

	outputChan := make(chan models.TransitData)

	if err := createTxPipe.AddDirective(outChanID, outputChan); err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		if routineErr := l1Oracle.EventLoop(); routineErr != nil {
			log.Printf("Error received from oracle event loop %e", err)
		}
	}()

	for td := range outputChan {
		log.Printf("===============================================")
		log.Printf("Received Contract creation Transaction %+v", td)
		log.Printf("===============================================")

		parsedTx, success := td.Value.(types.Transaction)
		if !success {
			log.Printf("Could not parse transaction value")
		}

		log.Printf("As parsed transaction %+v", parsedTx)
	}
}
