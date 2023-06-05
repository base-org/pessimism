package oracle

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
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

// TODO(#21): Verify config validity during Oracle construction
// EventDefintion ... Address register oracle definition used to drive oracle component
type EventDefintion struct {
	register   *core.DataRegister
	pUUID      core.PipelineUUID
	cfg        *core.OracleConfig
	client     client.EthClientInterface
	currHeight *big.Int
}

// NewEventDefintion ... Initializer for address.balance oracle definition
func NewEventDefintion(cfg *core.OracleConfig, client client.EthClientInterface, h *big.Int, reg *core.DataRegister) *EventDefintion {
	return &EventDefintion{
		register:   reg,
		cfg:        cfg,
		client:     client,
		currHeight: h,
	}
}

// NewAddressBalanceOracle ... Initializer for address.balance oracle component
func NewEventOracle(ctx context.Context, ot core.PipelineType,
	cfg *core.OracleConfig, reg *core.DataRegister, opts ...component.Option) (component.Component, error) {
	client := client.NewEthClient()

	od := NewEventDefintion(cfg, client, nil, reg)
	o, err := component.NewOracle(ctx, ot, core.GethBlock, od, opts...)
	if err != nil {
		return nil, err
	}

	return o, nil
}

// ConfigureRoutine ... Sets up the oracle client connection and persists puuid to definition state
func (oracle *EventDefintion) ConfigureRoutine(pUUID core.PipelineUUID) error {
	oracle.pUUID = pUUID

	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(core.EthClientTimeout))
	defer ctxCancel()

	logging.WithContext(ctxTimeout).Info("Setting up Account Balance client")

	return oracle.client.DialContext(ctxTimeout, oracle.cfg.RPCEndpoint)
}

// BackTestRoutine ...
// NOTE - This oracle does not support backtesting
// TODO (#59) : Add account balance backtesting support
func (oracle *EventDefintion) BackTestRoutine(_ context.Context, _ chan core.TransitData,
	_ *big.Int, _ *big.Int) error {
	return fmt.Errorf(noBackTestSupportError)
}

func (oracle *EventDefintion) backFill(startHeight, endHeight *big.Int, events []contractEvents) ([]types.Log, error) {
	var addresses []common.Address

	for _, event := range events {
		addresses = append(addresses, event.address)
	}

	query := ethereum.FilterQuery{
		FromBlock: startHeight,
		ToBlock:   endHeight,
		Addresses: addresses,
	}

	logs, err := oracle.client.FilterLogs(context.Background(), query)
	if err != nil {
		return []types.Log{}, err
	}

	var logEvents []types.Log
	for _, eLog := range logs {

		for _, event := range events {
			if event.address == eLog.Address && event.HasSignature(eLog.Topics[0]) {
				logEvents = append(logEvents, eLog)
			}
		}
	}

	return logEvents, nil
}

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

func (oracle *EventDefintion) getEventsToMonitor(ctx context.Context, addresses []string, ss state.Store) ([]contractEvents, error) {

	var events []contractEvents
	for _, address := range addresses {
		addressKey := state.MakeKey(core.EventPrefix, address, false).WithPUUID(oracle.pUUID)
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

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client for address (EOA, Contract) native balance amounts
func (oracle *EventDefintion) ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error {
	stateStore, err := state.FromContext(ctx)
	if err != nil {
		return err
	}

	if oracle.cfg.Backfill() { // Backfill routine
		addresses, err := stateStore.GetSlice(ctx, oracle.register.StateKeys[0].WithPUUID(oracle.pUUID))
		if err != nil {
			return err
		}

		events, err := oracle.getEventsToMonitor(ctx, addresses, stateStore)
		if err != nil {
			return err
		}

		oracle.backFill(oracle.cfg.StartHeight, oracle.cfg.EndHeight, events)
	}

	ticker := time.NewTicker(oracle.cfg.PollInterval * time.Millisecond) //nolint:durationcheck // inapplicable

	for {
		/*
			NOTE - This polling mechanism is not ideal given that read logic
			is ran on an interval and not on an event basis (ie. every block).
			Resulting in potential missed events or duplicate events.

			This is a temporary solution and will be fixed soon
			TODO(#60): Add Support for Oracle to Oracle Event Communication
		*/

		select {
		case <-ticker.C: // Polling
			logging.NoContext().Debug("Getting addresess",
				zap.String(core.PUUIDKey, oracle.pUUID.String()))

			addresses, err := stateStore.GetSlice(ctx, oracle.register.StateKeys[0].WithPUUID(oracle.pUUID))
			if err != nil {
				logging.WithContext(ctx).Error(err.Error())
				continue
			}

			eventsToMonitor, err := oracle.getEventsToMonitor(ctx, addresses, stateStore)
			if err != nil {
				logging.WithContext(ctx).Error(err.Error())
				continue
			}

			// Get event addresses from shared state store for pipeline uuid

			var addressesToMonitor []common.Address
			for _, address := range addresses {
				// Convert to go-ethereum address type
				gethAddress := common.HexToAddress(address)
				addressesToMonitor = append(addressesToMonitor, gethAddress)
			}

			recentBlock, err := oracle.client.BlockByNumber(context.Background(), nil)
			if err != nil {
				logging.WithContext(ctx).Error(err.Error())
				continue
			}

			hash := recentBlock.Header().Hash()
			query := ethereum.FilterQuery{
				BlockHash: &hash,
				Addresses: addressesToMonitor,
			}

			logging.WithContext(ctx).Debug("Getting logs")
			logs, err := oracle.client.FilterLogs(context.Background(), query)
			if err != nil {
				logging.WithContext(ctx).Error(err.Error())
				continue
			}

			logging.WithContext(ctx).Debug(fmt.Sprintf("got logs %+v", logs))

			if len(logs) == 0 {
				logging.WithContext(ctx).Debug("No logs found")
				continue
			}

			for _, log := range logs {
				for _, event := range eventsToMonitor {
					// Check if event is in the list of events to monitor
					if event.address == log.Address && event.HasSignature(log.Topics[0]) {
						componentChan <- core.NewTransitData(core.EventLog, log, core.WithAddress(log.Address))
					}
				}
			}

		case <-ctx.Done(): // Shutdown
			return nil
		}
	}
}
