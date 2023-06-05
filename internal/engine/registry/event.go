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
	Session UUID: %s
	Session Address: %s 
	Event: %s
	Transaction Hash: %s
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

	log, _ := td.Value.(types.Log)

	invalidated := true
	return &core.InvalOutcome{
		Message: fmt.Sprintf(eventReportMsg, bi.SUUID(), log.Address, string(log.Data), log.TxHash.Hex()),
	}, invalidated, nil
}
