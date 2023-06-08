package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/ethereum/go-ethereum/core/types"
)

// EventInvConfig  ... Configuration for the event invariant
type EventInvConfig struct {
	ContractName string   `json:"contract_name"`
	Address      string   `json:"address"`
	Args         []string `json:"args"`
}

// EventInvariant ...
type EventInvariant struct {
	cfg *EventInvConfig

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
	return &EventInvariant{
		cfg: cfg,

		Invariant: invariant.NewBaseInvariant(core.EventLog,
			invariant.WithAddressing()),
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

	return &core.InvalOutcome{
		Message: fmt.Sprintf(eventReportMsg, ei.cfg.ContractName, log.Address, log.TxHash.Hex(), ei.cfg.Args[0]),
	}, true, nil
}
