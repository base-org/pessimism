package registry

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/conduit/pipeline"
	"github.com/ethereum/go-ethereum/ethclient"
)

type (
	GethBlockOracleConfig struct {
		// EthRpc ... endpoint for geth API
		EthRpc string
		// Used for backfilling and backtesting
		StartHeight *int
		EndHeight   *int
	}

	GethBlockOracleDefinition struct {
		cfg *GethBlockOracleConfig

		// TODO - Bind this to an interface and mock so that this logic can be tested
		client *ethclient.Client

		// TODO - Represent with an enum
		evalFunc func(height int) bool
		backFill bool
		backTest bool
	}
)

func NewGethBlockOracle(cfg *GethBlockOracleConfig) pipeline.OracleDefinition {
	return &GethBlockOracleDefinition{cfg: cfg}

}

func (oracle *GethBlockOracleDefinition) backTestIncomplete(height int) bool {
	return height <= *oracle.cfg.EndHeight
}

func (oracle *GethBlockOracleDefinition) liveEval(height int) bool {
	return true
}

func (oracle *GethBlockOracleDefinition) ConfigureRoutine() error {
	// TODO - Introduce starting block parameter
	log.Print("Setting up client")
	client, err := ethclient.Dial(oracle.cfg.EthRpc)
	if err != nil {
		return err
	}

	oracle.client = client

	// Backtest
	if oracle.cfg.EndHeight != nil && oracle.cfg.StartHeight != nil {
		oracle.evalFunc = oracle.backTestIncomplete
	}

	return nil
}

func (oracle *GethBlockOracleDefinition) ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error {
	// NOTE - This poller logic is really bad and doesn't currently compensate for the following edge cases:
	// 1 - Client timeouts/failures; ie embed retry logic
	// 2 - Only reads most recent headers and doesn't take a starting block # - DONE
	// 3 - No optionality support for an ending block
	// 4 - Doesn't track live block value
	// 5 - No validation to ensure that blocks in monotonic height aren't ignored

	var height *big.Int = nil

	if oracle.cfg.StartHeight != nil {
		height = big.NewInt(int64(*oracle.cfg.StartHeight))
	}

	for {
		header, err := oracle.client.HeaderByNumber(ctx, height)
		if err != nil {
			log.Printf("Header fetching error: %s", err.Error())
			continue
		}

		block, err := oracle.client.BlockByNumber(ctx, header.Number)
		if err != nil {
			log.Printf("Error fetching block @ height %d: %s", height, err)
		}

		componentChan <- models.TransitData{
			Timestamp: time.Now(),
			Type:      "GETH_BLOCK",
			Value:     *block,
		}

		// height += 1
		height.Add(height, big.NewInt(1))
	}

}
