package registry

import (
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// BalanceInvConfig  ... Configuration for the balance invariant
type BalanceInvConfig struct {
	Address    string   `json:"address"`
	UpperBound *float64 `json:"upper"`
	LowerBound *float64 `json:"lower"`
}

// BalanceInvariant ...
type BalanceInvariant struct {
	cfg *BalanceInvConfig

	invariant.Invariant
}

// reportMsg ... Message to be sent to the alerting system
const reportMsg = `
	Current value: %3f
	Upper bound: %s
	Lower bound: %s

	Session UUID: %s
	Session Address: %s 
`

// NewBalanceInvariant ... Initializer
func NewBalanceInvariant(cfg *BalanceInvConfig) invariant.Invariant {
	return &BalanceInvariant{
		cfg: cfg,

		Invariant: invariant.NewBaseInvariant(core.AccountBalance),
	}
}

// Invalidate ... Checks if the balance is within the bounds
// specified in the config
func (bi *BalanceInvariant) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation for balance invariant", zap.String("data", fmt.Sprintf("%v", td)))

	if td.Type != bi.InputType() {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	balance, ok := td.Value.(float64)
	if !ok {
		return nil, false, fmt.Errorf("could not cast transit data value to float type")
	}

	invalidated := false

	// balance > upper bound
	if bi.cfg.UpperBound != nil &&
		*bi.cfg.UpperBound < balance {
		invalidated = true
	}

	// balance < lower bound
	if bi.cfg.LowerBound != nil &&
		*bi.cfg.LowerBound > balance {
		invalidated = true
	}

	if invalidated {
		var upper, lower string

		if bi.cfg.UpperBound != nil {
			upper = fmt.Sprintf("%2f", *bi.cfg.UpperBound)
		} else {
			upper = "∞"
		}

		if bi.cfg.LowerBound != nil {
			lower = fmt.Sprintf("%2f", *bi.cfg.LowerBound)
		} else {
			lower = "-∞"
		}

		return &core.InvalOutcome{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(reportMsg, balance,
				upper, lower,
				bi.SUUID(), bi.cfg.Address),
		}, true, nil
	}

	// No invalidation
	return nil, false, nil
}
