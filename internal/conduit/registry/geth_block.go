package registry

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/conduit/pipeline"
	"github.com/base-org/pessimism/internal/config"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	pollInterval = 200
)

// GethBlockODef ...GethBlock register oracle definition used to drive oracle component
type GethBlockODef struct {
	cfg *config.OracleConfig
	// TODO - Bind this to an interface and mock so that this logic can be tested
	client     *ethclient.Client
	currHeight *big.Int
}

// NewGethBlockOracle ... Initializer
func NewGethBlockOracle(ctx context.Context,
	ot pipeline.OracleType, cfg *config.OracleConfig) (pipeline.Component, error) {
	od := &GethBlockODef{cfg: cfg, currHeight: nil}

	return pipeline.NewOracle(ctx, ot, od)
}

func (oracle *GethBlockODef) ConfigureRoutine() error {
	log.Print("Setting up GETH Block client")

	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(models.EthClientTimeout))
	defer ctxCancel()

	client, err := ethclient.DialContext(ctxTimeout, oracle.cfg.RPCEndpoint)
	if err != nil {
		return err
	}

	oracle.client = client
	return nil
}

// BackTestRoutine ...
func (oracle *GethBlockODef) BackTestRoutine(_ context.Context, _ chan models.TransitData) error {
	// TODO - implement

	return nil
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client using monotonic block height variable for block metadata
// & writes block metadata to output listener components
func (oracle *GethBlockODef) ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error {
	// NOTE - Might need improvements in future as the project takes shape.

	ticker := time.NewTicker(pollInterval * time.Millisecond)

	for {
		select {
		case <-ticker.C:

			// How this works:
			// Check if current height is nil, if it is, then check if starting height is provided:
			// 1. if start height is provided, use that number as the current height
			// 2. if not, then sending nil as current height means use the latest
			// if current height is not nil, skip all above steps and continue iterating.
			// At the end, if the end height is specified and not nil, if its met, it returns once done.
			// Start Height and End Height is inclusive in fetching blocks.

			var height *big.Int

			if oracle.currHeight == nil {
				log.Printf("Current Height is nil, looking for starting height")
				if oracle.cfg.StartHeight != nil {
					log.Printf("StartHeight found to be: %d, using that value.", oracle.cfg.StartHeight)
					height = oracle.cfg.StartHeight
				} else {
					log.Printf("Starting Height is nil, using latest block as starting point.")
					height = nil
				}
			} else {
				height = oracle.currHeight
				log.Printf("Currently processing height: %d", height)
			}

			// retry for specified number of times.
			// Not an exponent backoff, but a simpler method which retries sooner
			var header *types.Header
			var err error

			// TODO: potentially break this functions off with the right context.
			for i := 1; i <= oracle.cfg.NumOfRetries; i++ {
				if i != 1 {
					log.Printf("Header Retry number: %d", i)
					time.Sleep(time.Duration(i) * time.Second)
				}

				header, err = oracle.client.HeaderByNumber(ctx, height)
				if err != nil {
					log.Printf("Header fetching error: %s", err.Error())
				}

				if i == oracle.cfg.NumOfRetries {
					log.Printf("All retries exhausted. Block at height %d will be skipped", height)
				}

			}

			// means the above retries failed
			if header == nil {
				continue
			}

			var block *types.Block

			// TODO: potentially break this functions off with the right context.
			for i := 1; i <= oracle.cfg.NumOfRetries; i++ {
				if i != 1 {
					log.Printf("Block Retry number: %d", i)
					time.Sleep(time.Duration(i) * time.Second)
				}

				block, err = oracle.client.BlockByNumber(ctx, header.Number)
				if err != nil {
					log.Printf("Block fetching error: %s", err.Error())
				}

				if i == oracle.cfg.NumOfRetries {
					log.Printf("All retries exhausted. Block at height %d will be skipped", height)
				}
			}

			// means the above retries failed
			if block == nil {
				continue
			}

			// TODO - Add support for database persistence
			componentChan <- models.TransitData{
				Timestamp: time.Now(),
				Type:      GethBlock,
				Value:     *block,
			}

			if height != nil {
				height.Add(height, big.NewInt(1))
			} else {
				height = header.Number
			}

			oracle.currHeight = height

			if oracle.cfg.EndHeight != nil && height == oracle.cfg.EndHeight {
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
