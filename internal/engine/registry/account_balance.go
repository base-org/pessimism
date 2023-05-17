package registry

import (
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/base-org/pessimism/internal/alert"
	pess_core "github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
)

const (
	lessThan    = -1
	greaterThan = 1
)

type BalanceInvConfig struct {
	Address    string   `json"address"`
	UpperBound *big.Int `json:"upper"`
	LowerBound *big.Int `json:"lower"`
}

type BalanceInvariant struct {
	cfg *BalanceInvConfig

	invariant.Invariant
}

var reportMsg = `
	Balance Enforcement Invalidation
	Current value: %d
	Upper bound: %d
	Lower bound: %d

	Session UUID: %s
	Session Address: %s 
`

func NewBalanceInvariant(cfg *BalanceInvConfig) invariant.Invariant {
	return &BalanceInvariant{
		cfg: cfg,

		Invariant: invariant.NewBaseInvariant(pess_core.AccountBalance, invariant.WithAddressing()),
	}
}

func (bi *BalanceInvariant) Invalidate(td pess_core.TransitData) (bool, error) {
	logging.NoContext().Debug("Checking invalidation")

	if td.Type != bi.InputType() {
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

	url := os.Getenv("SLACK_URL") // TODO - Pass down from config
	alert.SlackHandler(
		fmt.Sprintf(reportMsg, balance,
			bi.cfg.UpperBound, bi.cfg.LowerBound,
			bi.UUID(), bi.cfg.Address),
		http.Client{},
		url,
	)

	return invalidated, nil
}
