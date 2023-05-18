package registry

import (
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/core"
	pess_core "github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
)

const (
	lessThan    = -1
	greaterThan = 1
)

// BalanceInvConfig  ...
type BalanceInvConfig struct {
	Address    string   `json:"address"`
	UpperBound *big.Int `json:"upper"`
	LowerBound *big.Int `json:"lower"`
}

// BalanceInvariant ...
type BalanceInvariant struct {
	cfg *BalanceInvConfig

	invariant.Invariant
}

const reportMsg = `
	Current value: %d
	Upper bound: %d
	Lower bound: %d

	Session UUID: %s
	Session Address: %s 
`

// NewBalanceInvariant ... Initializer
func NewBalanceInvariant(cfg *BalanceInvConfig) invariant.Invariant {
	return &BalanceInvariant{
		cfg: cfg,

		Invariant: invariant.NewBaseInvariant(pess_core.AccountBalance, invariant.WithAddressing()),
	}
}

// Invalidate ... Checks if the balance is within the bounds
// specified in the config
func (bi *BalanceInvariant) Invalidate(td pess_core.TransitData) (*core.InvalOutcome, error) {
	logging.NoContext().Debug("Checking invalidation")

	if td.Type != bi.InputType() {
		return nil, fmt.Errorf("invalid type supplied")
	}

	balance, ok := td.Value.(*big.Int)
	if !ok || balance == nil {
		return nil, fmt.Errorf("could not cast transit data value to int type")
	}

	invalidated := false

	// balance > upper bound
	if bi.cfg.UpperBound != nil &&
		bi.cfg.UpperBound.Cmp(balance) == lessThan {
		invalidated = true
	}

	// balance < lower bound
	if bi.cfg.LowerBound != nil &&
		bi.cfg.LowerBound.Cmp(balance) == greaterThan {
		invalidated = true
	}

	if invalidated {
		return &core.InvalOutcome{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(reportMsg, balance,
				bi.cfg.UpperBound, bi.cfg.LowerBound,
				bi.UUID(), bi.cfg.Address),
			SUUID: bi.UUID(),
		}, nil
	}

	// No invalidation
	return nil, nil
}
