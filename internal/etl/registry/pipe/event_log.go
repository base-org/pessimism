package pipe

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	pess_common "github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"go.uber.org/zap"
)

// EventDefinition ...
type EventDefinition struct {
	client client.EthClientInterface
	sk     *core.StateKey
	pUUID  core.PUUID
	cfg    *core.ClientConfig
	ss     state.Store
}

// NewEventParserPipe ... Initializer
func NewEventParserPipe(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	client, err := client.FromContext(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	ss, err := state.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	ed := &EventDefinition{
		cfg:    cfg,
		client: client,
		ss:     ss,
	}

	p, err := component.NewPipe(ctx, ed, core.GethBlock, core.EventLog, opts...)
	if err != nil {
		return nil, err
	}

	ed.sk = p.StateKey().Clone()
	return p, nil
}

// getEventsToMonitor ... Gets the smart contract events to monitor from the state store
func (ed *EventDefinition) getTopics(ctx context.Context,
	addresses []string, ss state.Store) [][]common.Hash {
	events := make([]common.Hash, 0)

	for _, address := range addresses {
		innerKey := &core.StateKey{
			Nesting: false,
			Prefix:  ed.sk.Prefix,
			ID:      address,
			PUUID:   &ed.pUUID,
		}

		sigs, err := ss.GetSlice(ctx, innerKey)
		if err != nil {
			logging.WithContext(ctx).Error(err.Error())
		}

		for _, sig := range sigs {
			events = append(events, crypto.Keccak256Hash([]byte(sig)))
		}
	}

	topics := make([][]common.Hash, 1)
	topics[0] = events

	return topics
}

// Transform ... Gets the events from the block, filters them and
// returns them if they are in the list of events to monitor
func (ed *EventDefinition) Transform(ctx context.Context, td core.TransitData) ([]core.TransitData, error) {
	// 1. Convert arbitrary transit data to go-ethereum compatible block type
	block, success := td.Value.(types.Block)
	if !success {
		return []core.TransitData{}, fmt.Errorf("could not convert to block")
	}

	// 2. Fetch the addresess and events to monitor for
	logging.NoContext().Debug("Getting addresess",
		zap.String(core.PUUIDKey, ed.pUUID.String()))

	addresses, err := ed.ss.GetSlice(ctx, ed.sk)
	if err != nil {
		return []core.TransitData{}, err
	}

	topics := ed.getTopics(ctx, addresses, ed.ss)
	hash := block.Header().Hash()

	// 3. Construct and execute a filter query on the provided block
	// to get the relevant logs
	query := ethereum.FilterQuery{
		BlockHash: &hash,
		Addresses: pess_common.SliceToAddresses(addresses),
		Topics:    topics,
	}

	logs, err := ed.client.FilterLogs(context.Background(), query)
	if err != nil {
		return []core.TransitData{}, err
	}

	// 4. See if there are any logs to process
	if len(logs) == 0 {
		return []core.TransitData{}, nil
	}

	// 5. Convert the logs to transit data and return them
	result := make([]core.TransitData, 0)
	for _, log := range logs {
		result = append(result,
			core.NewTransitData(core.EventLog, log, core.WithAddress(log.Address)))
	}

	return result, nil
}
