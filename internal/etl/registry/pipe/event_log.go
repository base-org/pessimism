package pipe

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	p_common "github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"go.uber.org/zap"
)

const (
	// dlqMaxSize ... Max size of the DLQ
	// NOTE ... This could be made configurable via env vars
	// or some other mechanism if needed
	dlqMaxSize = 100
)

// EventDefinition ... Represents the stateful definition of the event log pipe component
// Used to transform block data into event logs (block->events)
// Uses a DLQ to reprocess failed queries & state store to get events to monitor
type EventDefinition struct {
	client client.EthClient
	dlq    *p_common.DLQ[core.TransitData]

	pUUID core.PUUID
	ss    state.Store

	SK *core.StateKey
}

// NewEventDefinition ... Initializes the event log pipe definition
func NewEventDefinition(ctx context.Context, n core.Network) (*EventDefinition, error) {
	// 1. Load dependencies from context
	client, err := client.FromContext(ctx, n)
	if err != nil {
		return nil, err
	}

	ss, err := state.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Construct the pipe definition
	ed := &EventDefinition{
		dlq:    p_common.NewTransitDLQ(dlqMaxSize),
		client: client,
		ss:     ss,
	}
	return ed, nil
}
<<<<<<< HEAD

// NewEventParserPipe ... Initializer
func NewEventParserPipe(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	// 1. Construct the pipe definition
	ed, err := NewEventDefinition(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

=======

// NewEventParserPipe ... Initializer
func NewEventParserPipe(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	// 1. Construct the pipe definition
	ed, err := NewEventDefinition(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

>>>>>>> d9da3aec27e458936600f46e40aed8141688994d
	// 2. Embed the definition into a generic pipe construction
	p, err := component.NewPipe(ctx, ed, core.GethBlock, core.EventLog, opts...)
	if err != nil {
		return nil, err
	}

	// 3. Set the post component construction fields on the definition
	// There's likely a more extensible way to construct this definition fields
	// given that they're used by component implementations across the ETL
	ed.SK = p.StateKey().Clone()
	ed.pUUID = p.PUUID()
	return p, nil
}

// getEventsToMonitor ... Gets the smart contract events to monitor from the state store
func (ed *EventDefinition) getTopics(ctx context.Context,
	addresses []string, ss state.Store) [][]common.Hash {
	sigs := make([]common.Hash, 0)

	// 1. Iterate over addresses and construct nested state keys
	// to lookup the associated events to monitor
	for _, address := range addresses {
		innerKey := &core.StateKey{
			Nesting: false,
			Prefix:  ed.SK.Prefix,
			ID:      address,
			PUUID:   ed.SK.PUUID,
		}

		// 1.1 Attempt to fetch the events to monitor from the state store
		// and continue if there is an error
		events, err := ss.GetSlice(ctx, innerKey)
		if err != nil {
			logging.WithContext(ctx).Error("Failed to get events to monitor",
				zap.String(logging.PUUIDKey, ed.pUUID.String()),
				zap.Error(err))
			continue
		}

		// 1.2 Compute signatures for the function declaration strings
		for _, event := range events {
			sigs = append(sigs, crypto.Keccak256Hash([]byte(event)))
		}
	}

	// 2. Construct the topics slice to be used in the filter query
	// via populating the first index of 2D Topics slice with the event signatures to monitor
	topics := make([][]common.Hash, 1)
	topics[0] = sigs

	return topics
}

// Transform ... Attempts to reprocess previously failed queries first
// before attempting to process the current block data
func (ed *EventDefinition) Transform(ctx context.Context, td core.TransitData) ([]core.TransitData, error) {
	logger := logging.WithContext(ctx)
	// 1. Check to see if there are any failed queries to reprocess
	// If failures occur again, add the caller (Transform)
	// function input to the DLQ and return
	var (
		tds []core.TransitData
		err error
	)

	if !ed.dlq.Empty() {
		logger.Debug("Attempting to reprocess failed queries",
			zap.Int("dlq_size", ed.dlq.Size()))

		tds, err = ed.attemptDLQ(ctx)
		// NOTE ... Returning here is intentional to ensure that block events
		// downstream are processed in the sequential order for which they came in
		if err != nil {
			err = ed.dlq.Add(&td)
			if err != nil {
				return tds, err
			}
		}
		logger.Debug("Successfully reprocessed failed queries",
			zap.String(logging.PUUIDKey, ed.pUUID.String()))
	}

	// 2. If there are no failed queries, then process the current block data
	// and add a data input to the DLQ if it fails for reprocessing next block
	tds2, err := ed.transformFunc(ctx, td)
	if err != nil {
		if ed.dlq.Full() {
			// NOTE ... If the DLQ is full, then we pop the oldest entry
			// to make room for the new entry
			lostVal, _ := ed.dlq.Pop()
			logger.Warn("DLQ is full, popping oldest entry",
				zap.String(logging.PUUIDKey, ed.pUUID.String()),
				zap.Any("lost_value", lostVal))

			metrics.WithContext(ctx).
				IncMissedBlock(ed.pUUID)
		}

		// NOTE ... If the DLQ is not full, then we add the entry to the DLQ
		_ = ed.dlq.Add(&td)
		logging.WithContext(ctx).Error("Failed to process block data",
			zap.Int("dlq_size", ed.dlq.Size()))

		return tds, err
	}

	// 3. Concatenate the results from the failed queries and the current block data
	// and return
	tds = append(tds, tds2...)
	return tds, nil
}

// attemptDLQ ... Attempts to reprocess previously failed queries
func (ed *EventDefinition) attemptDLQ(ctx context.Context) ([]core.TransitData, error) {
	failedInputs := ed.dlq.PopAll()

	// 1. Attempt to reprocess failed inputs
	tds := make([]core.TransitData, 0)
	for _, td := range failedInputs {
		result, err := ed.transformFunc(ctx, *td)
		// 2. If the reprocessing fails, then the function will return an error
		if err != nil {
			err = ed.dlq.Add(td)
			if err != nil {
				return tds, err
			}
			// NOTE ... Returning here is intentional to ensure that block events
			// downstream are processed in the sequential order for which they came in
			return tds, err
		}

		// 3. If the reprocessing succeeds, append result to return slice
		tds = append(tds, result...)
	}

	return tds, nil
}

// transformFunc ... Gets the events from the block, filters them and
// returns them if they are in the list of events to monitor
func (ed *EventDefinition) transformFunc(ctx context.Context, td core.TransitData) ([]core.TransitData, error) {
	// 1. Convert arbitrary transit data to go-ethereum compatible block type
	block, success := td.Value.(types.Block)
	if !success {
		return []core.TransitData{}, fmt.Errorf("could not convert to block")
	}

	// 2. Fetch the addresses and events to monitor for
	logging.NoContext().Debug("Getting addresses",
		zap.String(logging.PUUIDKey, ed.pUUID.String()))

	addresses, err := ed.ss.GetSlice(ctx, ed.SK)
	if err != nil {
		return []core.TransitData{}, err
	}

	topics := ed.getTopics(ctx, addresses, ed.ss)
	hash := block.Header().Hash()

	// 3. Construct and execute a filter query on the provided block hash
	// to get the relevant logs
	query := ethereum.FilterQuery{
		BlockHash: &hash,
		Addresses: p_common.SliceToAddresses(addresses),
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
			core.NewTransitData(core.EventLog, log, core.WithAddress(log.Address),
				core.WithOriginTS(td.OriginTS)))
	}

	return result, nil
}
