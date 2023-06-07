package pipe

import (
	"context"
	"fmt"
	"time"

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
	sk     core.StateKey
	pUUID  core.PipelineUUID
	cfg    *core.ClientConfig
}

// ConfigureRoutine ... Sets up the oracle client connection and persists puuid to definition state
func (ed *EventDefinition) ConfigureRoutine(pUUID core.PipelineUUID) error {
	ed.pUUID = pUUID
	ed.sk = state.MakeKey(core.EventLog, core.AddressKey, true).
		WithPUUID(pUUID)

	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(core.EthClientTimeout))
	defer ctxCancel()

	logging.WithContext(ctxTimeout).Info("Setting up GETH client connection")

	err := ed.client.DialContext(ctxTimeout, ed.cfg.RPCEndpoint)

	if err != nil {
		return err
	}
	return nil

}

// NewEventParserPipe ... Initializer
func NewEventParserPipe(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	client := client.NewEthClient()

	ed := &EventDefinition{
		cfg:    cfg,
		client: client,
	}

	return component.NewPipe(ctx, ed, core.GethBlock, core.EventLog, opts...)
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
func (ed *EventDefinition) getEventsToMonitor(ctx context.Context, rt core.RegisterType,
	addresses []string, ss state.Store) ([]contractEvents, error) {
	var events []contractEvents
	for _, address := range addresses {
		addressKey := state.MakeKey(rt, address, false).WithPUUID(ed.pUUID)
		sigs, err := ss.GetSlice(ctx, addressKey)
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

// Transform ... Gets the events from the block, filters them and returns them if they are in the list of events to monitor
func (ed *EventDefinition) Transform(ctx context.Context, td core.TransitData) ([]core.TransitData, error) {
	asBlock, success := td.Value.(types.Block)
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

	eventsToMonitor, err := ed.getEventsToMonitor(ctx, core.EventLog, addresses, stateStore)
	if err != nil {
		return []core.TransitData{}, err
	}

	hash := asBlock.Header().Hash()

	query := ethereum.FilterQuery{
		BlockHash: &hash,
		Addresses: pess_common.SliceToAddresses(addresses),
	}

	logs, err := ed.client.FilterLogs(context.Background(), query)

	returnVals := make([]core.TransitData, 0)
	for _, log := range logs {
		for _, event := range eventsToMonitor {
			// Check if event is in the list of events to monitor
			if event.address == log.Address && event.HasSignature(log.Topics[0]) {
				returnVals = append(returnVals,
					core.NewTransitData(core.EventLog, log, core.WithAddress(log.Address)))
			}
		}
	}

	return returnVals, nil
}
