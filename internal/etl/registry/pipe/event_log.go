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
	client client.EthClient
	Sk     core.StateKey
	pUUID  core.PipelineUUID
	cfg    *core.ClientConfig
}

// ConfigureRoutine ... Sets up the pipe client connection and persists puuid to definition state
func (ed *EventDefinition) ConfigureRoutine(pUUID core.PipelineUUID) error {
	ed.pUUID = pUUID
	// TODO(#69): State Key Representation is Insecure
	ed.Sk = state.MakeKey(core.EventLog, core.AddressKey, true).
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

// NewEventDefinition ... Initializer for pipe definition
func NewEventDefinition(cfg *core.ClientConfig, client client.EthClient) *EventDefinition {
	return &EventDefinition{
		cfg:    cfg,
		client: client,
	}
}

// NewEventParserPipe ... Initializer
func NewEventParserPipe(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	ec := client.NewEthClient()
	ed := NewEventDefinition(cfg, ec)

	return component.NewPipe(ctx, ed, core.GethBlock, core.EventLog, opts...)
}

// getEventsToMonitor ... Gets the smart contract events to monitor from the state store
func (ed *EventDefinition) getEventsToMonitor(ctx context.Context, rt core.RegisterType,
	addresses []string, ss state.Store) (map[common.Address][]common.Hash, error) {
	eventMap := make(map[common.Address][]common.Hash)

	for _, address := range addresses {
		addrKey := state.MakeKey(rt, address, false).WithPUUID(ed.pUUID)
		sigs, err := ss.GetSlice(ctx, addrKey)
		if err != nil {
			logging.WithContext(ctx).Error(err.Error())
			return nil, err
		}

		var parsedSigs []common.Hash
		for _, sig := range sigs {
			hash := crypto.Keccak256Hash([]byte(sig))
			parsedSigs = append(parsedSigs, hash)
		}

		logging.WithContext(ctx).Debug(fmt.Sprintf("Address: %s, Sigs: %v", address, parsedSigs))

		gethAddr := common.HexToAddress(address)
		eventMap[gethAddr] = parsedSigs
	}

	return eventMap, nil
}

// Transform ... Gets the events from the block, filters them and
// returns them if they are in the list of events to monitor
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

	addresses, err := stateStore.GetSlice(ctx, ed.Sk)
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
	if err != nil {
		return []core.TransitData{}, err
	}

	result := make([]core.TransitData, 0)
	for _, log := range logs {
		if _, exists := eventsToMonitor[log.Address]; !exists {
			continue
		}

		for _, eventSig := range eventsToMonitor[log.Address] {
			if eventSig == log.Topics[0] {
				result = append(result,
					core.NewTransitData(core.EventLog, log, core.WithAddress(log.Address)),
				)
			}
		}
	}

	return result, nil
}
