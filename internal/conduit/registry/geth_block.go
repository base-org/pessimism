package registry

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/conduit/pipeline"
	"github.com/base-org/pessimism/internal/config"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	pollInterval = 200
)

// GethBlockODef ...GethBlock register oracle definition used to drive oracle component
type GethBlockODef struct {
	cfg        *config.OracleConfig
	client     client.EthClientInterface
	currHeight *big.Int
}

// NewGethBlockOracle ... Initializer
func NewGethBlockOracle(ctx context.Context,
	ot pipeline.OracleType, cfg *config.OracleConfig, client client.EthClientInterface) (pipeline.Component, error) {
	od := &GethBlockODef{cfg: cfg, currHeight: nil, client: client}
	return pipeline.NewOracle(ctx, ot, od)
}

func (oracle *GethBlockODef) ConfigureRoutine() error {
	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(models.EthClientTimeout))
	defer ctxCancel()

	log := ctxzap.Extract(ctxTimeout)
	log.Info("Setting up GETH Block client")

	err := oracle.client.DialContext(ctxTimeout, oracle.cfg.RPCEndpoint)

	if err != nil {
		return err
	}
	return nil
}

// getCurrentHeightFromNetwork ... Gets the current height of the network and will not quit until found
func (oracle *GethBlockODef) getCurrentHeightFromNetwork(ctx context.Context) *types.Header {
	log := ctxzap.Extract(ctx)

	for {
		header, err := oracle.client.HeaderByNumber(ctx, nil)
		if err != nil {
			log.Error("problem fetching current height from network", zap.Error(err))
			continue
		}
		return header
	}
}

// BackTestRoutine ...
func (oracle *GethBlockODef) BackTestRoutine(ctx context.Context, componentChan chan models.TransitData,
	startHeight *big.Int, endHeight *big.Int) error {
	log := ctxzap.Extract(ctx)
	if endHeight.Cmp(startHeight) < 0 {
		return errors.New("start height cannot be more than the end height")
	}

	currentHeader := oracle.getCurrentHeightFromNetwork(ctx)

	if startHeight.Cmp(currentHeader.Number) == 1 {
		return errors.New("start height cannot be more than the latest height from network")
	}

	ticker := time.NewTicker(pollInterval * time.Millisecond)
	height := startHeight

	for {
		select {
		case <-ticker.C:

			headerAsInterface, err := oracle.fetchData(ctx, height, models.FetchHeader)
			headerAsserted, headerAssertedOk := headerAsInterface.(*types.Header)
			// means the above retries failed
			if err != nil || !headerAssertedOk {
				log.Error("problem fetching header", zap.Error(err))

				continue
			}

			blockAsInterface, err := oracle.fetchData(ctx, headerAsserted.Number, models.FetchBlock)
			blockAsserted, blockAssertedOk := blockAsInterface.(*types.Block)

			if err != nil || !blockAssertedOk {
				log.Error("problem fetching block", zap.Error(err))
				continue
			}

			// TODO - Add support for database persistence
			componentChan <- models.TransitData{
				Timestamp: time.Now(),
				Type:      GethBlock,
				Value:     *blockAsserted,
			}

			if height.Cmp(endHeight) == 0 {
				log.Info("Completed back-test routine.")
				return nil
			}

			height.Add(height, big.NewInt(1))

		case <-ctx.Done():
			return nil
		}
	}
}

// getHeightToProcess ...
//
//	Check if current height is nil, if it is, then check if starting height is provided:
//	1. if start height is provided, use that number as the current height
//	2. if not, then sending nil as current height means use the latest
//	if current height is not nil, skip all above steps and continue iterating.
//	At the end, if the end height is specified and not nil, if its met, it returns once done.
//	Start Height and End Height is inclusive in fetching blocks.
func (oracle *GethBlockODef) getHeightToProcess(ctx context.Context) *big.Int {
	log := ctxzap.Extract(ctx)
	if oracle.currHeight == nil {
		log.Info("Current Height is nil, looking for starting height")
		if oracle.cfg.StartHeight != nil {
			log.Info("StartHeight found to be: %d, using that value.", zap.Int64("StartHeight",
				oracle.cfg.StartHeight.Int64()))
			return oracle.cfg.StartHeight
		}
		log.Info("Starting Height is nil, using latest block as starting point.")
		return nil
	}
	return oracle.currHeight
}

// fetchHeaderWithRetry ... retry for specified number of times.
// Not an exponent backoff, but a simpler method which retries sooner
func (oracle *GethBlockODef) fetchData(ctx context.Context, height *big.Int,
	fetchType models.FetchType) (interface{}, error) {
	if fetchType == models.FetchHeader {
		return oracle.client.HeaderByNumber(ctx, height)
	}
	return oracle.client.BlockByNumber(ctx, height)
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client using monotonic block height variable for block metadata
// & writes block metadata to output listener components
func (oracle *GethBlockODef) ReadRoutine(ctx context.Context, componentChan chan models.TransitData) error {
	// NOTE - Might need improvements in future as the project takes shape.
	log := ctxzap.Extract(ctx)

	if oracle.cfg.EndHeight != nil && oracle.cfg.StartHeight == nil {
		return errors.New("cannot start with latest block height with end height configured")
	}

	if oracle.cfg.EndHeight.Cmp(oracle.cfg.StartHeight) < 0 {
		return errors.New("start height cannot be more than the end height")
	}

	// Now fetching current height from the network
	currentHeader := oracle.getCurrentHeightFromNetwork(ctx)

	if oracle.cfg.StartHeight.Cmp(currentHeader.Number) == 1 {
		return errors.New("start height cannot be more than the latest height from network")
	}

	ticker := time.NewTicker(pollInterval * time.Millisecond)
	for {
		select {
		case <-ticker.C:

			height := oracle.getHeightToProcess(ctx)

			headerAsInterface, err := oracle.fetchData(ctx, height, models.FetchHeader)
			headerAsserted, headerAssertedOk := headerAsInterface.(*types.Header)
			// means the above retries failed
			if err != nil || !headerAssertedOk {
				log.Error("problem fetching header", zap.Error(err))
				continue
			}

			blockAsInterface, err := oracle.fetchData(ctx, headerAsserted.Number, models.FetchBlock)
			blockAsserted, blockAssertedOk := blockAsInterface.(*types.Block)

			if err != nil || !blockAssertedOk {
				log.Error("problem fetching block %v", zap.Error(err))
				continue
			}

			// TODO - Add support for database persistence
			componentChan <- models.TransitData{
				Timestamp: time.Now(),
				Type:      GethBlock,
				Value:     *blockAsserted,
			}

			// check has to be done here to include the end height block
			if oracle.cfg.EndHeight != nil && height.Cmp(oracle.cfg.EndHeight) == 0 {
				return nil
			}

			if height != nil {
				height.Add(height, big.NewInt(1))
			} else {
				height = &big.Int{}
				height.Add(headerAsserted.Number, big.NewInt(1))
			}

			oracle.currHeight = height

		case <-ctx.Done():
			return nil
		}
	}
}
