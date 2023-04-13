package registry

import (
	"context"
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/conduit/pipeline"
	"github.com/base-org/pessimism/internal/config"
)

const (
	pollInterval = 200
)

// GethBlockODef ...GethBlock register oracle definition used to drive oracle component
type GethBlockODef struct {
	cfg             *config.OracleConfig
	clientInterface pipeline.EthClientInterface
	currHeight      *big.Int
}

// NewGethBlockOracle ... Initializer
func NewGethBlockOracle(ctx context.Context,
	ot pipeline.OracleType, cfg *config.OracleConfig, clientNew pipeline.EthClientInterface) (pipeline.Component, error) {
	od := &GethBlockODef{cfg: cfg, currHeight: nil, clientInterface: clientNew}
	return pipeline.NewOracle(ctx, ot, od)
}

func (oracle *GethBlockODef) ConfigureRoutine() error {
	log.Print("Setting up GETH Block client")

	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(models.EthClientTimeout))
	defer ctxCancel()

	err := oracle.clientInterface.DialContext(ctxTimeout, oracle.cfg.RPCEndpoint)

	if err != nil {
		return err
	}

	return nil
}

// BackTestRoutine ...
func (oracle *GethBlockODef) BackTestRoutine(ctx context.Context, componentChan chan models.TransitData,
	startHeight *big.Int, endHeight *big.Int) error {
	if endHeight.Cmp(startHeight) < 0 {
		return errors.New("start height cannot be more than the end height")
	}

	ticker := time.NewTicker(pollInterval * time.Millisecond)
	height := startHeight

	for {
		select {
		case <-ticker.C:

			header := oracle.fetchHeaderWithRetry(ctx, height)

			// means the above retries failed, and keep trying
			if header == nil {
				continue
			}

			block := oracle.fetchHeaderWithRetry(ctx, header.Number)

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

			if height.Cmp(endHeight) == 0 {
				log.Printf("Completed backtest routine.")
				return nil
			}

			height.Add(height, big.NewInt(1))

		case <-ctx.Done():
			return nil
		}
	}
}

func (oracle *GethBlockODef) getHeightToProcess() *big.Int {
	if oracle.currHeight == nil {
		log.Printf("Current Height is nil, looking for starting height")
		if oracle.cfg.StartHeight != nil {
			log.Printf("StartHeight found to be: %d, using that value.", oracle.cfg.StartHeight)
			return oracle.cfg.StartHeight
		}

		log.Printf("Starting Height is nil, using latest block as starting point.")
		return nil
	}

	log.Printf("Currently processing height: %d", oracle.currHeight)
	return oracle.currHeight
}

// fetchHeaderWithRetry ... retry for specified number of times.
// Not an exponent backoff, but a simpler method which retries sooner
func (oracle *GethBlockODef) fetchHeaderWithRetry(ctx context.Context, height *big.Int) *types.Header {
	for i := 0; i <= oracle.cfg.NumOfRetries; i++ {
		if i != 0 {
			log.Printf("Header Retry number: %d", i)
			time.Sleep(time.Duration(i) * time.Second)
		}

		header, err := oracle.clientInterface.HeaderByNumber(ctx, height)
		if err != nil {
			log.Printf("Header fetching error: %s", err.Error())
		} else {
			return header
		}

		if i == oracle.cfg.NumOfRetries {
			log.Printf("All retries exhausted. Block at height %d will be skipped", height)
			return nil
		}
	}
	return nil
}

// fetchBlockWithRetry ... retry for specified number of times.
// Not an exponent backoff, but a simpler method which retries sooner
func (oracle *GethBlockODef) fetchBlockWithRetry(ctx context.Context, headerNumber *big.Int) *types.Block {
	for i := 0; i <= oracle.cfg.NumOfRetries; i++ {
		if i != 0 {
			log.Printf("Block Retry number: %d", i)
			time.Sleep(time.Duration(i) * time.Second)
		}

		block, err := oracle.clientInterface.BlockByNumber(ctx, headerNumber)
		if err != nil {
			log.Printf("Block fetching error: %s", err.Error())
		} else {
			return block
		}

		if i == oracle.cfg.NumOfRetries {
			log.Printf("All retries exhausted. Block at height %d will be skipped", headerNumber)
			return nil
		}
	}
	return nil
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client using monotonic block height variable for block metadata
// & writes block metadata to output listener components
func (oracle *GethBlockODef) ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error {
	// NOTE - Might need improvements in future as the project takes shape.

	if oracle.cfg.EndHeight != nil && oracle.cfg.StartHeight == nil {
		return errors.New("cannot start with latest block height with end height configured")
	}

	if oracle.cfg.EndHeight.Cmp(oracle.cfg.StartHeight) < 0 {
		return errors.New("start height cannot be more than the end height")
	}

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

			height := oracle.getHeightToProcess()

			// TODO: potentially break this functions off with the right context.
			header := oracle.fetchHeaderWithRetry(ctx, height)

			// means the above retries failed
			if header == nil {
				continue
			}

			block := oracle.fetchBlockWithRetry(ctx, header.Number)

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

			// check has to be done here to include the end height block
			if oracle.cfg.EndHeight != nil && height.Cmp(oracle.cfg.EndHeight) == 0 {
				return nil
			}

			if height != nil {
				height.Add(height, big.NewInt(1))
			} else {
				height = &big.Int{}
				height.Add(header.Number, big.NewInt(1))
			}

			oracle.currHeight = height

		case <-ctx.Done():
			return nil
		}
	}
}
