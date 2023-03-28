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

	gethConfig := &registry.GethBlockOracleConfig{
		EthRpc:      cfg.L1RpcEndpoint,
		StartHeight: start,
		EndHeight:   nil}

	l1OracleDef := registry.NewGethBlockOracle(gethConfig)

	id := xid.New()
	outChan := make(chan models.TransitData)

	router := pipeline.NewOutputRouter()
	router.AddDirective(id, outChan)

	l1BlockOracle := pipeline.NewOracle(appCtx, pipeline.Live, l1OracleDef)
	l1BlockOracle.AddDirective(id, outChan)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		if err := l1BlockOracle.EventLoop(); err != nil {
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
