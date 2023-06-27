package oracle

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

const (
	notFoundMsg = "not found"
)

// TODO(#21): Verify config validity during Oracle construction
// GethBlockODef ...GethBlock register oracle definition used to drive oracle component
type GethBlockODef struct {
	cUUID      core.CUUID
	cfg        *core.ClientConfig
	client     client.EthClientInterface
	currHeight *big.Int
}

// NewGethBlockODef ... Initializer for geth.block oracle definition
func NewGethBlockODef(cfg *core.ClientConfig, client client.EthClientInterface, h *big.Int) *GethBlockODef {
	return &GethBlockODef{
		cfg:        cfg,
		client:     client,
		currHeight: h,
	}
}

// NewGethBlockOracle ... Initializer for geth.block oracle component
func NewGethBlockOracle(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	client, err := client.FromContext(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	od := NewGethBlockODef(cfg, client, nil)

	oracle, err := component.NewOracle(ctx, core.GethBlock, od, opts...)
	if err != nil {
		return nil, err
	}

	od.cUUID = oracle.UUID()
	return oracle, nil
}

// ConfigureRoutine ... Sets up the oracle client connection and persists puuid to definition state
func (oracle *GethBlockODef) ConfigureRoutine(core.PUUID) error {
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

	ticker := time.NewTicker(oracle.cfg.PollInterval * time.Millisecond) //nolint:durationcheck // inapplicable
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
				// logging.WithContext(ctx).Error("problem fetching or asserting block", zap.NamedError("blockFetch", err),
				// 	zap.Bool("blockAsserted", blockAssertedOk))
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

func validHeightParams(start, end *big.Int) error {
	if end != nil && start == nil {
		return errors.New("cannot start with latest block height with end height configured")
	}

	if end != nil && start != nil &&
		end.Cmp(start) < 0 {
		return errors.New("start height cannot be more than the end height")
	}

	return nil
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client using monotonic block height variable for block metadata
// & writes block metadata to output listener components
func (oracle *GethBlockODef) ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error {
	// NOTE - Might need improvements in future as the project takes shape.

	// Now fetching current height from the network
	// currentHeader := oracle.getCurrentHeightFromNetwork(ctx)

	// if oracle.cfg.StartHeight.Cmp(currentHeader.Number) == 1 {
	// 	return errors.New("start height cannot be more than the latest height from network")
	// }

	if err := validHeightParams(oracle.cfg.StartHeight, oracle.cfg.EndHeight); err != nil {
		return err
	}

	logging.WithContext(ctx).
		Debug("Starting poll routine", zap.Duration("poll_interval", oracle.cfg.PollInterval),
			zap.String(core.CUUIDKey, oracle.cUUID.String()))

	ticker := time.NewTicker(oracle.cfg.PollInterval * time.Millisecond) //nolint:durationcheck // inapplicable
	for {
		select {
		case <-ticker.C:

			height := oracle.getHeightToProcess(ctx)
			if height != nil {
				logging.WithContext(ctx).Debug("Polling block for processing",
					zap.Int("Height", int(height.Int64())),
					zap.String(core.CUUIDKey, oracle.cUUID.String()))
			}

			headerAsInterface, err := oracle.fetchData(ctx, height, core.FetchHeader)
			headerAsserted, headerAssertedOk := headerAsInterface.(*types.Header)

			if err != nil && err.Error() == notFoundMsg {
				continue
			}

			if err != nil || !headerAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting header", zap.NamedError("headerFetch", err),
					zap.Bool("headerAsserted", headerAssertedOk), zap.String(core.CUUIDKey, oracle.cUUID.String()))
				continue
			}

			blockAsInterface, err := oracle.fetchData(ctx, headerAsserted.Number, core.FetchBlock)
			blockAsserted, blockAssertedOk := blockAsInterface.(*types.Block)

			if err != nil || !blockAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting block", zap.NamedError("blockFetch", err),
					zap.Bool("blockAsserted", blockAssertedOk), zap.String(core.CUUIDKey, oracle.cUUID.String()))
				continue
			}

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

			logging.NoContext().Debug("New height", zap.Int("Height", int(height.Int64())),
				zap.String(core.CUUIDKey, oracle.cUUID.String()))

			oracle.currHeight = height

		case <-ctx.Done():
			logging.NoContext().Info("Geth.block oracle routine ending", zap.String(core.CUUIDKey, oracle.cUUID.String()))
			return nil
		}
	}
}
