package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	p_common "github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum-optimism/optimism/op-service/retry"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"go.uber.org/zap"
)

// LogSubscription ...
type LogSubscription struct {
	client client.EthClient

	PathID core.PathID
	ss     state.Store

	SK *core.StateKey
}

// logSub ...
func logSub(ctx context.Context, n core.Network) (*LogSubscription, error) {
	client, err := client.FromNetwork(ctx, n)
	if err != nil {
		return nil, err
	}

	ss, err := state.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	lspt := &LogSubscription{
		client: client,
		ss:     ss,
	}
	return lspt, nil
}

// NewLogSubscriber ... Initializer
func NewLogSubscriber(ctx context.Context, cfg *core.ClientConfig,
	opts ...process.Option) (process.Process, error) {
	// 1. Construct the pipe definition
	s, err := logSub(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	// 2. Embed the definition into a generic pipe construction
	p, err := process.NewSubscriber(ctx, s, core.BlockHeader, core.Log, opts...)
	if err != nil {
		return nil, err
	}

	// 3. Set the post component construction fields on the definition
	// There's likely a more extensible way to construct this definition fields
	// given that they're used by component implementations across the ETL
	s.SK = p.StateKey().Clone()
	s.PathID = p.PathID()
	return p, nil
}

// getEventsToMonitor ... Gets the smart contract events to monitor from the state store
func (sub *LogSubscription) getTopics(ctx context.Context,
	addresses []string, ss state.Store) [][]common.Hash {
	sigs := make([]common.Hash, 0)

	// 1. Iterate over addresses and construct nested state keys
	// to lookup the associated events to monitor
	for _, address := range addresses {
		innerKey := &core.StateKey{
			Nesting: false,
			Prefix:  sub.SK.Prefix,
			ID:      address,
			PathID:  sub.SK.PathID,
		}

		// 1.1 Attempt to fetch the events to monitor from the state store
		// and continue if there is an error
		events, err := ss.GetSlice(ctx, innerKey)
		if err != nil {
			logging.WithContext(ctx).Error("Failed to get events to monitor",
				zap.String(logging.PathIDKey, sub.PathID.String()),
				zap.Error(err))
			continue
		}

		// 1.2 Compute signatures for the function declaration strings
		for _, event := range events {
			sigs = append(sigs, crypto.Keccak256Hash([]byte(event)))
		}
	}

	// populate event signatures to monitor
	topics := make([][]common.Hash, 1)
	topics[0] = sigs

	return topics
}

func (sub LogSubscription) Run(ctx context.Context, e core.Event) ([]core.Event, error) {
	logger := logging.WithContext(ctx)
	tds, err := sub.transformFunc(ctx, e)
	if err != nil {
		logger.Error("Failed to process block data",
			zap.String(logging.PathIDKey, sub.PathID.String()),
			zap.Error(err))

		return nil, err
	}

	return tds, nil
}

func (sub *LogSubscription) transformFunc(ctx context.Context, e core.Event) ([]core.Event, error) {
	header, success := e.Value.(types.Header)
	if !success {
		return []core.Event{}, fmt.Errorf("could not convert to header")
	}

	logging.NoContext().Debug("Getting addresses",
		zap.String(logging.PathIDKey, sub.PathID.String()))

	addresses, err := sub.ss.GetSlice(ctx, sub.SK)
	if err != nil {
		return []core.Event{}, err
	}

	topics := sub.getTopics(ctx, addresses, sub.ss)
	hash := header.Hash()

	// Construct and execute a filter query on the provided block hash
	// to get the relevant logs
	query := ethereum.FilterQuery{
		BlockHash: &hash,
		Addresses: p_common.SliceToAddresses(addresses),
		Topics:    topics,
	}

	logs, err := retry.Do[[]types.Log](ctx, 10, core.RetryStrategy(), func() ([]types.Log, error) {
		return sub.client.FilterLogs(context.Background(), query)
	})

	if err != nil {
		logging.WithContext(ctx).Error("Failed to parse transform events", zap.Error(err))
	}

	if len(logs) == 0 {
		return []core.Event{}, nil
	}

	result := make([]core.Event, 0)
	for _, log := range logs {
		result = append(result,
			core.NewEvent(core.Log, log, core.WithAddress(log.Address),
				core.WithOriginTS(e.OriginTS)))
	}

	return result, nil
}
