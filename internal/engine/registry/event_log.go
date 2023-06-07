package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// EventInvConfig  ...
type EventInvConfig struct {
	Address string   `json:"address"`
	Args    []string `json:"args"`
}

// EventInvariant ...
type EventInvariant struct {
	cfg *EventInvConfig

	invariant.Invariant
}

// eventReportMsg ... Message to be sent to the alerting system
const eventReportMsg = `
	_Monitored Event Triggered_

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
func (bi *EventInvariant) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation for balance invariant", zap.String("data", fmt.Sprintf("%v", td)))

	if td.Type != bi.InputType() {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	return &core.InvalOutcome{
		Message: fmt.Sprintf(eventReportMsg, log.Address, log.TxHash.Hex(), bi.cfg.Args[0]),
	}, true, nil
}
