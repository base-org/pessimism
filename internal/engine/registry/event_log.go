package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// EventInvConfig  ... Configuration for the event invariant
type EventInvConfig struct {
	ContractName string   `json:"contract_name"`
	Address      string   `json:"address"`
	Sigs         []string `json:"args"`
}

// EventInvariant ...
type EventInvariant struct {
	cfg  *EventInvConfig
	sigs []common.Hash

	invariant.Invariant
}

// eventReportMsg ... Message to be sent to the alerting system
const eventReportMsg = `
	_Monitored Event Triggered_

	Contract Name: %s
	Contract Address: %s 
	Transaction Hash: %s
	Event: %s
`

// NewEventInvariant ... Initializer
func NewEventInvariant(cfg *EventInvConfig) invariant.Invariant {
	var sigs []common.Hash
	for _, sig := range cfg.Sigs {
		sigs = append(sigs, crypto.Keccak256Hash([]byte(sig)))
	}

	return &EventInvariant{
		cfg:  cfg,
		sigs: sigs,

		Invariant: invariant.NewBaseInvariant(core.EventLog),
	}
}

// Invalidate ... Checks if the balance is within the bounds
// specified in the config
func (ei *EventInvariant) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	if td.Type != ei.InputType() {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	if td.Address.String() != ei.cfg.Address {
		return nil, false, fmt.Errorf("invalid address supplied")
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, false, fmt.Errorf("could not convert transit data to log")
	}

	var invalidated = false

	for _, sig := range ei.sigs {
		if log.Topics[0] == sig {
			invalidated = true
			break
		}
	}

	if !invalidated {
		return nil, false, nil
	}

	return &core.InvalOutcome{
		Message: fmt.Sprintf(eventReportMsg, ei.cfg.ContractName, log.Address, log.TxHash.Hex(), ei.cfg.Sigs[0]),
	}, true, nil
}
