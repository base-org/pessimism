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
		RpcEndpoint: cfg.L1RpcEndpoint,
		StartHeight: nil,
		EndHeight:   nil}

	// 1. Configure blackhole tx pipe component
	createRegister, err := registry.GetRegister(registry.CONTRACT_CREATE_TX)
	if err != nil {
		panic(err)
	}

	initPipe, success := createRegister.ComponentConstructor.(pipeline.PipeConstructorFunc)

	inputChan := make(chan models.TransitData)

	createTxPipe := initPipe(appCtx, inputChan)

	register, err := registry.GetRegister(registry.GETH_BLOCK)
	if err != nil {
		panic(err)
	}

	init, success := register.ComponentConstructor.(pipeline.OracleConstructor)
	if !success {
		panic("Could not read constructor value")
	}

	go func() {
		if err := createTxPipe.EventLoop(); err != nil {
			log.Printf("Error recieved from oracle event loop %e", err)
		}
		return
	}()

	l1Oracle := init(appCtx, pipeline.LiveOracle, l1OracleCfg)

	l1Oracle.AddDirective(0x42, inputChan)

	outputChan := make(chan models.TransitData)
	createTxPipe.AddDirective(0x420, outputChan)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		if err := l1Oracle.EventLoop(); err != nil {
			log.Printf("Error recieved from oracle event loop %e", err)
		}
		return
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
