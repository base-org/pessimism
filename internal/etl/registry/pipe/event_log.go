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
}

// ConfigureRoutine ... Sets up the pipe client connection and persists puuid to definition state
func (ed *EventDefinition) ConfigureRoutine(pUUID core.PUUID) error {
	ed.pUUID = pUUID
	return nil
}

// NewEventParserPipe ... Initializer
func NewEventParserPipe(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	client, err := client.FromContext(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	ed := &EventDefinition{
		cfg:    cfg,
		client: client,
	}

	p, err := component.NewPipe(ctx, ed, core.GethBlock, core.EventLog, opts...)
	if err != nil {
		return nil, err
	}

	ed.sk = p.StateKey().Clone()
	return p, nil
}

// contractEvents ... Struct to hold the contract address and the event signatures
type contractEvents struct {
	address common.Address
	sigs    []common.Hash
}

// HasSignature ... Checks if the event has the signature
func (ce *contractEvents) HasSignature(sig common.Hash) bool {
	for _, s := range ce.sigs {
		if s == sig {
			return true
		}
	}

	return false
}

// getEventsToMonitor ... Gets the smart contract events to monitor from the state store
func (ed *EventDefinition) getEventsToMonitor(ctx context.Context,
	addresses []string, ss state.Store) ([]contractEvents, error) {
	var events []contractEvents
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
			return []contractEvents{}, err
		}

		var parsedSigs []common.Hash
		for _, sig := range sigs {
			parsedSigs = append(parsedSigs, crypto.Keccak256Hash([]byte(sig)))
		}

		logging.WithContext(ctx).Debug(fmt.Sprintf("Address: %s, Sigs: %v", address, parsedSigs))
		events = append(events, contractEvents{
			address: common.HexToAddress(address),
			sigs:    parsedSigs,
		})
	}

	return events, nil
}

// Transform ... Gets the events from the block, filters them and
// returns them if they are in the list of events to monitor
func (ed *EventDefinition) Transform(ctx context.Context, td core.TransitData) ([]core.TransitData, error) {
	block, success := td.Value.(types.Block)
	if !success {
		return []core.TransitData{}, fmt.Errorf("could not convert to block")
	}

	stateStore, err := state.FromContext(ctx)
	if err != nil {
		return []core.TransitData{}, err
	}

	logging.NoContext().Debug("Getting addresess",
		zap.String(core.PUUIDKey, ed.pUUID.String()))

	addresses, err := stateStore.GetSlice(ctx, ed.sk)
	if err != nil {
		return []core.TransitData{}, err
	}

	eventsToMonitor, err := ed.getEventsToMonitor(ctx, addresses, stateStore)
	if err != nil {
		return []core.TransitData{}, err
	}

	hash := block.Header().Hash()

	query := ethereum.FilterQuery{
		BlockHash: &hash,
		Addresses: pess_common.SliceToAddresses(addresses),
	}

	logs, err := ed.client.FilterLogs(context.Background(), query)
	if err != nil {
		return []core.TransitData{}, err
	}

	result := make([]core.TransitData, 0)
	for _, log := range logs {
		for _, event := range eventsToMonitor {
			// Check if event is in the list of events to monitor
			if event.address == log.Address && event.HasSignature(log.Topics[0]) {
				result = append(result,
					core.NewTransitData(core.EventLog, log, core.WithAddress(log.Address)))
			}
		}
	}

	return result, nil
}
