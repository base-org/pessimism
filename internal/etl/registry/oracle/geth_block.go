package oracle

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"strconv"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
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
	pUUID      core.PUUID
	cfg        *core.ClientConfig
	client     client.EthClient
	rpcClient  *rpc.Client
	currHeight *big.Int

	stats metrics.Metricer
}

// NewGethBlockODef ... Initializer for geth.block oracle definition
func NewGethBlockODef(cfg *core.ClientConfig, client client.EthClient, rawClient *rpc.Client,
	h *big.Int, stats metrics.Metricer) *GethBlockODef {
	return &GethBlockODef{
		cfg:        cfg,
		client:     client,
		rpcClient:  rawClient,
		currHeight: h,
		stats:      stats,
	}
}

// NewGethBlockOracle ... Initializer for geth.block oracle component
func NewGethBlockOracle(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	ethClient, err := client.FromContext(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	var rawClient *rpc.Client
	if cfg.Network == core.Layer1 {
		rawClient = client.GetRawL1Client(ctx)
	} else if cfg.Network == core.Layer2 {
		rawClient = client.GetRawL2Client(ctx)
	}

	od := NewGethBlockODef(cfg, ethClient, rawClient, nil, metrics.WithContext(ctx))

	oracle, err := component.NewOracle(ctx, core.GethBlock, od, opts...)
	if err != nil {
		return nil, err
	}

	od.cUUID = oracle.UUID()
	od.pUUID = oracle.PUUID()
	return oracle, nil
}

// getCurrentHeightFromNetwork ... Gets the current height of the network and will not quit until found
func (oracle *GethBlockODef) getCurrentHeightFromNetwork(ctx context.Context) *types.Header {
	for {
		header, err := oracle.client.HeaderByNumber(ctx, nil)
		if err != nil {
			oracle.stats.RecordNodeError(oracle.cfg.Network)
			logging.WithContext(ctx).Error("problem fetching current height from network", zap.Error(err))
			continue
		}
		return header
	}
}

// BlockByRange non-inclusive.
func (oracle *GethBlockODef) BlockByRange(ctx context.Context, startHeight, endHeight *big.Int) ([]*types.Header, error) {
	count := new(big.Int).Sub(endHeight, startHeight).Uint64()
	batchElems := make([]rpc.BatchElem, count)
	for i := uint64(0); i < count; i++ {
		height := new(big.Int).Add(startHeight, new(big.Int).SetUint64(i))
		batchElems[i] = rpc.BatchElem{
			Method: "eth_getBlockByNumber",
			Args:   []interface{}{toBlockNumArg(height), false},
			Result: new(types.Header),
			Error:  nil,
		}
	}

	err := oracle.rpcClient.BatchCallContext(ctx, batchElems)
	if err != nil {
		return nil, err
	}

	// Parse the headers.
	//  - Ensure integrity that they build on top of each other
	//  - Truncate out headers that do not exist (endHeight > "latest")
	size := 0
	blocks := make([]*types.Header, count)
	for i, batchElem := range batchElems {
		if batchElem.Error != nil {
			return nil, batchElem.Error
		} else if batchElem.Result == nil {
			break
		}

		block := batchElem.Result.(*types.Header)

		blocks[i] = block
		size = size + 1
	}
	blocks = blocks[:size]

	return blocks, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	} else if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}

	// It's negative.
	if number.IsInt64() {
		tag, _ := rpc.BlockNumber(number.Int64()).MarshalText()
		return string(tag)
	}

	// It's negative and large, which is invalid.
	return fmt.Sprintf("<invalid %d>", number)
}

func timer() func() {
	start := time.Now()
	return func() {
		logging.NoContext().Info("took", zap.String("took", time.Since(start).String()))
	}
}

func (oracle *GethBlockODef) BatchedSendBlocks(ctx context.Context, startHeight *big.Int,
	endHeight *big.Int, componentChan chan core.TransitData) {
	defer timer()
	totalBlocks := new(big.Int).Sub(endHeight, startHeight).Uint64() + 1
	numOfBatches := totalBlocks / 1000 // already floor division
	processingStartHeight := startHeight
	for i := uint64(0); i < numOfBatches; i++ {
		logging.WithContext(ctx).Info("batch processing blocks", zap.String("batch",
			strconv.FormatUint(i, 10)))
		processingEndHeight := new(big.Int).Add(processingStartHeight, big.NewInt(1000))
		headers, err := oracle.BlockByRange(ctx, processingStartHeight, processingEndHeight)

		if err != nil {
			logging.WithContext(ctx).Error("problem fetching batch of blocks",
				zap.NamedError("batchBlockFetch", err))
			oracle.stats.RecordNodeError(oracle.cfg.Network)
			i-- // added this so that it can retry.
		}

		for _, eachHeader := range headers {
			componentChan <- core.TransitData{
				OriginTS:  time.Now(),
				Timestamp: time.Now(),
				Type:      core.GethHeader,
				Value:     *eachHeader,
			}
		}
		processingStartHeight = processingEndHeight
		logging.WithContext(ctx).Info("processed batch", zap.String("batch", strconv.FormatUint(i, 10)),
			zap.String("newStartHeight", processingStartHeight.String()))
	}

	remainingBlocks := new(big.Int).SetUint64(1 + totalBlocks - (numOfBatches * 1000))
	blocks, err := oracle.BlockByRange(ctx,
		new(big.Int).Mul(new(big.Int).SetUint64(numOfBatches), big.NewInt(1000)), remainingBlocks)

	logging.WithContext(ctx).Info("remaining blocks", zap.Uint64("blocks", remainingBlocks.Uint64()))

	if err != nil {
		logging.WithContext(ctx).Error("problem fetching batch of blocks", zap.NamedError("batchBlockFetch", err))
		oracle.stats.RecordNodeError(oracle.cfg.Network)
	}

	for _, eachHeader := range blocks {
		logging.WithContext(ctx).Info("eachHeader", zap.String("header", eachHeader.Number.String()))

		componentChan <- core.TransitData{
			OriginTS:  time.Now(),
			Timestamp: time.Now(),
			Type:      core.GethHeader,
			Value:     *eachHeader,
		}
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

	totalBlocks := new(big.Int).Sub(endHeight, startHeight).Uint64() + 1

	if totalBlocks > 1000 {
		logging.WithContext(ctx).Info("getting blocks in batches")
		oracle.BatchedSendBlocks(ctx, startHeight, endHeight, componentChan)
	} else {

		blocks, err := oracle.BlockByRange(ctx, startHeight, new(big.Int).Add(endHeight, big.NewInt(1)))
		if err != nil {
			logging.WithContext(ctx).Error("problem fetching batch of blocks", zap.NamedError("batchBlockFetch", err))
			oracle.stats.RecordNodeError(oracle.cfg.Network)
		}

		for _, eachBlock := range blocks {
			componentChan <- core.TransitData{
				OriginTS:  time.Now(),
				Timestamp: time.Now(),
				Type:      core.GethBlock,
				Value:     *eachBlock,
			}
		}
	}

	for {
		select {
		case <-ticker.C:

			headerAsInterface, err := oracle.fetchData(ctx, height, core.FetchHeader)
			headerAsserted, headerAssertedOk := headerAsInterface.(*types.Header)

			if err != nil || !headerAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting header", zap.NamedError("headerFetch", err),
					zap.Bool("headerAsserted", headerAssertedOk))
				oracle.stats.RecordNodeError(oracle.cfg.Network)
				continue
			}

			blockAsInterface, err := oracle.fetchData(ctx, headerAsserted.Number, core.FetchBlock)
			blockAsserted, blockAssertedOk := blockAsInterface.(*types.Block)

			if err != nil || !blockAssertedOk {
				// logging.WithContext(ctx).Error("problem fetching or asserting block", zap.NamedError("blockFetch", err),
				// 	zap.Bool("blockAsserted", blockAssertedOk))
				oracle.stats.RecordNodeError(oracle.cfg.Network)
				continue
			}

			// TODO - Add support for database persistence
			componentChan <- core.TransitData{
				OriginTS:  time.Now(),
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
			logging.WithContext(ctx).Info("using this value as start height", zap.Int64("StartHeight",
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
			zap.String(logging.CUUIDKey, oracle.cUUID.String()))

	if oracle.cfg.StartHeight != nil {
		logging.WithContext(ctx).Info("start height", zap.String("startHeight", oracle.cfg.StartHeight.String()))
		currentMaxHeight := oracle.getCurrentHeightFromNetwork(ctx)
		oracle.BatchedSendBlocks(ctx, big.NewInt(1000000), currentMaxHeight.Number, componentChan)
	}

	ticker := time.NewTicker(oracle.cfg.PollInterval * time.Millisecond) //nolint:durationcheck // inapplicable
	for {
		select {
		case <-ticker.C:
			opStart := time.Now()

			height := oracle.getHeightToProcess(ctx)
			if height != nil {
				logging.WithContext(ctx).Debug("Polling block for processing",
					zap.Int("Height", int(height.Int64())),
					zap.String(logging.CUUIDKey, oracle.cUUID.String()))
			}

			headerAsInterface, err := oracle.fetchData(ctx, height, core.FetchHeader)
			headerAsserted, headerAssertedOk := headerAsInterface.(*types.Header)

			// Ensure err is indicative of block not existing yet
			if err != nil && err.Error() == notFoundMsg {
				continue
			}

			if err != nil || !headerAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting header", zap.NamedError("headerFetch", err),
					zap.Bool("headerAsserted", headerAssertedOk), zap.String(logging.CUUIDKey, oracle.cUUID.String()))
				oracle.stats.RecordNodeError(oracle.cfg.Network)
				continue
			}

			blockAsInterface, err := oracle.fetchData(ctx, headerAsserted.Number, core.FetchBlock)
			block, blockAssertedOk := blockAsInterface.(*types.Block)

			if err != nil || !blockAssertedOk {
				logging.WithContext(ctx).Error("problem fetching or asserting block", zap.NamedError("blockFetch", err),
					zap.Bool("blockAsserted", blockAssertedOk), zap.String(logging.CUUIDKey, oracle.cUUID.String()))
				oracle.stats.RecordNodeError(oracle.cfg.Network)
				continue
			}

			blockTS := time.Unix(int64(block.Time()), 0)
			oracle.stats.RecordBlockLatency(oracle.cfg.Network, float64(time.Since(blockTS).Milliseconds()))

			componentChan <- core.TransitData{
				OriginTS:  opStart,
				Timestamp: time.Now(),
				Type:      core.GethBlock,
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
				height.Add(headerAsserted.Number, big.NewInt(1))
			}

			logging.NoContext().Debug("New height", zap.Int("Height", int(height.Int64())),
				zap.String(logging.CUUIDKey, oracle.cUUID.String()))

			oracle.currHeight = height

		case <-ctx.Done():
			logging.NoContext().Info("Geth.block oracle routine ending", zap.String(logging.CUUIDKey, oracle.cUUID.String()))
			return nil
		}
	}
}
