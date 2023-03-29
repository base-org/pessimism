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
	"github.com/rs/xid"
)

func main() {

	appCtx := context.Background()
	cfg := config.NewConfig("config.env")

	// TODO - Consider making this value not a pointer
	var start *int
	start = new(int)
	*start = 0

	l1OracleCfg := &config.OracleConfig{
		RpcEndpoint: cfg.L1RpcEndpoint,
		StartHeight: start,
		EndHeight:   nil}

	register, err := registry.GetRegister(registry.GETH_BLOCK)
	if err != nil {
		panic(err)
	}

	init, success := register.ComponentConstructor.(pipeline.OracleConstructor)
	if !success {
		panic("Could not read constructor value")
	}

	l1Oracle := init(appCtx, pipeline.LiveOracle, l1OracleCfg)

	id := xid.New()
	outChan := make(chan models.TransitData)

	router := pipeline.NewOutputRouter()
	router.AddDirective(id, outChan)

	l1Oracle.AddDirective(id, outChan)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		if err := l1Oracle.EventLoop(); err != nil {
			log.Printf("Error recieved from oracle event loop %e", err)
		}
		return
	}()

	for td := range outChan {
		log.Printf("Received transit data %+v", td)
		parsedBlock, success := td.Value.(types.Block)
		if !success {
			log.Printf("Could not parse block value")
		}

		log.Printf("(%d) As parsed block %+v", parsedBlock.Number(), parsedBlock)
	}
}
