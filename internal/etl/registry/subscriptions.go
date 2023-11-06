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
	PathID core.PathID
	SK     *core.StateKey

	client client.EthClient
	ss     state.Store
}

// NewLogSubscript ...
func NewLogSubscript(ctx context.Context, n core.Network) (*LogSubscription, error) {
	client, err := client.FromNetwork(ctx, n)
	if err != nil {
		return nil, err
	}

	ss, err := state.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	sub := &LogSubscription{
		client: client,
		ss:     ss,
	}
	return sub, nil
}

// NewLogSubscriber ... Initializer
func NewLogSubscriber(ctx context.Context, cfg *core.ClientConfig,
	opts ...process.Option) (process.Process, error) {
	s, err := NewLogSubscript(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	p, err := process.NewSubscriber(ctx, s, core.BlockHeader, core.Log, opts...)
	if err != nil {
		return nil, err
	}

	s.SK = p.StateKey().Clone()
	s.PathID = p.PathID()
	return p, nil
}

// Fetch smart contract events to monitor
func (sub *LogSubscription) getTopics(ctx context.Context,
	addresses []string, ss state.Store) [][]common.Hash {
	sigs := make([]common.Hash, 0)

	for _, address := range addresses {
		innerKey := &core.StateKey{
			Nesting: false,
			Prefix:  sub.SK.Prefix,
			ID:      address,
			PathID:  sub.SK.PathID,
		}

		events, err := ss.GetSlice(ctx, innerKey)
		if err != nil {
			logging.WithContext(ctx).Error("Failed to get events to monitor",
				zap.String(logging.Path, sub.PathID.String()),
				zap.Error(err))
			continue
		}

		for _, event := range events {
			sigs = append(sigs, crypto.Keccak256Hash([]byte(event)))
		}
	}

	// populate event signatures to monitor
	topics := make([][]common.Hash, 1)
	topics[0] = sigs

	return topics
}

func (sub *LogSubscription) Run(ctx context.Context, e core.Event) ([]core.Event, error) {
	logger := logging.WithContext(ctx)
	events, err := sub.transformEvents(ctx, e)
	if err != nil {
		logger.Error("Failed to process block data",
			zap.String(logging.Path, sub.PathID.String()),
			zap.Error(err))

		return nil, err
	}

	return events, nil
}

func (sub *LogSubscription) transformEvents(ctx context.Context, e core.Event) ([]core.Event, error) {
	header, success := e.Value.(types.Header)
	if !success {
		return []core.Event{}, fmt.Errorf("could not convert to header")
	}

	logging.NoContext().Debug("Getting addresses",
		zap.String(logging.Path, sub.PathID.String()))

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
		return []core.Event{}, err
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
