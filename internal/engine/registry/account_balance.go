package registry

import (
	"fmt"
	"math/big"

	pess_core "github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
)

const (
	lessThan    = -1
	greaterThan = 1
)

type BalanceInvConfig struct {
	UpperBound *big.Int `json:"upper"`
	LowerBound *big.Int `json:"lower"`
}

type BalanceInvariant struct {
	cfg *BalanceInvConfig

	invariant.Invariant
}

func NewBalanceInvariant(cfg *BalanceInvConfig) invariant.Invariant {
	return &BalanceInvariant{
		cfg: cfg,

		Invariant: invariant.NewBaseInvariant(pess_core.ContractCreateTX),
	}
}

func (bi *BalanceInvariant) InputType() pess_core.RegisterType {
	return pess_core.AccountBalance
}

func (bi *BalanceInvariant) Invalidate(td pess_core.TransitData) (bool, error) {
	logging.NoContext().Debug("Checking invalidation")

	if td.Type != pess_core.AccountBalance {
		return false, fmt.Errorf("invalid type supplied")
	}

	balance, ok := td.Value.(*big.Int)
	if !ok || balance == nil {
		return false, fmt.Errorf("could not cast transit data value to int type")
	}

	invalidated := false

	if bi.cfg.UpperBound != nil &&
		bi.cfg.UpperBound.Cmp(balance) == lessThan {
		invalidated = true
	}

	if bi.cfg.LowerBound != nil &&
		bi.cfg.LowerBound.Cmp(balance) == greaterThan {
		invalidated = true
	}

	return invalidated, nil
}
