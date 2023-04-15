package registry

import (
	"context"
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

const (
	pollInterval = 1000
)

// TODO(#21): Verify config validity during Oracle construction
// GethBlockODef ...GethBlock register oracle definition used to drive oracle component
type GethBlockODef struct {
	cfg        *config.OracleConfig
	client     client.EthClientInterface
	currHeight *big.Int
}

// NewGethBlockOracle ... Initializer
func NewGethBlockOracle(ctx context.Context, ot core.PipelineType,
	cfg *config.OracleConfig, opts ...component.Option) (component.Component, error) {

	client := client.NewEthClient()

	od := &GethBlockODef{cfg: cfg, currHeight: nil, client: client}
	return component.NewOracle(ctx, ot, core.GethBlock, od, opts...)
}

func (oracle *GethBlockODef) ConfigureRoutine() error {
	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(core.EthClientTimeout))
	defer ctxCancel()

	logging.WithContext(ctxTimeout).Info("Setting up GETH Block client")

	err := oracle.client.DialContext(ctxTimeout, oracle.cfg.RPCEndpoint)

	if err != nil {
		return err
	}
	return nil
}

// getCurrentHeightFromNetwork ... Gets the current height of the network and will not quit until found
func (oracle *GethBlockODef) getCurrentHeightFromNetwork(ctx context.Context) *types.Header {
	for {
		header, err := oracle.client.HeaderByNumber(ctx, nil)
		if err != nil {
			logging.WithContext(ctx).Error("problem fetching current height from network", zap.Error(err))
			continue
		}
		return header
	}
}

// BackTestRoutine ...
func (oracle *GethBlockODef) BackTestRoutine(ctx context.Context, componentChan chan core.TransitData,
	startHeight *big.Int, endHeight *big.Int) error {
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

			headerAsInterface, err := oracle.fetchData(ctx, height, core.FetchHeader)
			headerAsserted, headerAssertedOk := headerAsInterface.(*types.Header)

			if err != nil || !headerAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting header", zap.NamedError("headerFetch", err),
					zap.Bool("headerAsserted", headerAssertedOk))
				continue
			}

			blockAsInterface, err := oracle.fetchData(ctx, headerAsserted.Number, core.FetchBlock)
			blockAsserted, blockAssertedOk := blockAsInterface.(*types.Block)

			if err != nil || !blockAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting block", zap.NamedError("blockFetch", err),
					zap.Bool("blockAsserted", blockAssertedOk))
				continue
			}

			// TODO - Add support for database persistence
			componentChan <- core.TransitData{
				Timestamp: time.Now(),
				Type:      core.GethBlock,
				Value:     *blockAsserted,
			}

			if height.Cmp(endHeight) == 0 {
				logging.WithContext(ctx).Info("Completed back-test routine.")
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
	if oracle.currHeight == nil {
		logging.WithContext(ctx).Info("Current Height is nil, looking for starting height")
		if oracle.cfg.StartHeight != nil {
			logging.WithContext(ctx).Info("StartHeight found to be: %d, using that value.", zap.Int64("StartHeight",
				oracle.cfg.StartHeight.Int64()))
			return oracle.cfg.StartHeight
		}
		logging.WithContext(ctx).Info("Starting Height is nil, using latest block as starting point.")
		return nil
	}
	return oracle.currHeight
}

// fetchHeaderWithRetry ... retry for specified number of times.
// Not an exponent backoff, but a simpler method which retries sooner
func (oracle *GethBlockODef) fetchData(ctx context.Context, height *big.Int,
	fetchType core.FetchType) (interface{}, error) {
	if fetchType == core.FetchHeader {
		return oracle.client.HeaderByNumber(ctx, height)
	}
	return oracle.client.BlockByNumber(ctx, height)
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client using monotonic block height variable for block metadata
// & writes block metadata to output listener components
func (oracle *GethBlockODef) ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error {
	// NOTE - Might need improvements in future as the project takes shape.

	if oracle.cfg.EndHeight != nil && oracle.cfg.StartHeight == nil {
		return errors.New("cannot start with latest block height with end height configured")
	}

	if oracle.cfg.EndHeight.Cmp(oracle.cfg.StartHeight) < 0 {
		return errors.New("start height cannot be more than the end height")
	}

	// Now fetching current height from the network
	// currentHeader := oracle.getCurrentHeightFromNetwork(ctx)

	// if oracle.cfg.StartHeight.Cmp(currentHeader.Number) == 1 {
	// 	return errors.New("start height cannot be more than the latest height from network")
	// }

	ticker := time.NewTicker(pollInterval * time.Millisecond)
	for {
		select {
		case <-ticker.C:

			height := oracle.getHeightToProcess(ctx)

			headerAsInterface, err := oracle.fetchData(ctx, height, core.FetchHeader)
			headerAsserted, headerAssertedOk := headerAsInterface.(*types.Header)

			if err != nil || !headerAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting header", zap.NamedError("headerFetch", err),
					zap.Bool("headerAsserted", headerAssertedOk))
				continue
			}

			blockAsInterface, err := oracle.fetchData(ctx, headerAsserted.Number, core.FetchBlock)
			blockAsserted, blockAssertedOk := blockAsInterface.(*types.Block)

			if err != nil || !blockAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting block", zap.NamedError("blockFetch", err),
					zap.Bool("blockAsserted", blockAssertedOk))
				continue
			}

			log.Printf("%d", height)

			// TODO - Add support for database persistence
			componentChan <- core.TransitData{
				Timestamp: time.Now(),
				Type:      core.GethBlock,
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
