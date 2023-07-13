package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// BalanceInvConfig ... Configuration for the balance invariant
type BalanceInvConfig struct {
	Address    string   `json:"address"`
	UpperBound *float64 `json:"upper"`
	LowerBound *float64 `json:"lower"`
}

// Unmarshal ... Converts a general config to a balance invariant config
func (bi *BalanceInvConfig) Unmarshal(isp *core.InvSessionParams) error {
	return json.Unmarshal(isp.Bytes(), &bi)
}

// BalanceInvariant ...
type BalanceInvariant struct {
	cfg *BalanceInvConfig

	invariant.Invariant
}

// reportMsg ... Message to be sent to the alerting subsystem
const reportMsg = `
	Current value: %3f
	Upper bound: %s
	Lower bound: %s

	Session UUID: %s
	Session Address: %s 
`

// NewBalanceInvariant ... Initializer
func NewBalanceInvariant(cfg *BalanceInvConfig) (invariant.Invariant, error) {
	return &BalanceInvariant{
		cfg:       cfg,
		Invariant: invariant.NewBaseInvariant(core.AccountBalance),
	}, nil
}

// Invalidate ... Checks if the balance is within the bounds
// specified in the config
func (bi *BalanceInvariant) Invalidate(td core.TransitData) (*core.Invalidation, bool, error) {
	logging.NoContext().Debug("Checking invalidation for balance invariant", zap.String("data", fmt.Sprintf("%v", td)))

	// 1. Validate and extract balance input
	err := bi.ValidateInput(td)
	if err != nil {
		return nil, false, err
	}

	balance, ok := td.Value.(float64)
	if !ok {
		return nil, false, fmt.Errorf(couldNotCastErr, "float64")
	}

	invalidated := false

	// 2. Invalidate if balance > upper bound
	if bi.cfg.UpperBound != nil &&
		*bi.cfg.UpperBound < balance {
		invalidated = true
	}

	// 3. Invalidate if balance < lower bound
	if bi.cfg.LowerBound != nil &&
		*bi.cfg.LowerBound > balance {
		invalidated = true
	}

	/// 4. Generate invalidation outcome if invalidated
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

		return &core.Invalidation{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(reportMsg, balance,
				upper, lower,
				bi.SUUID(), bi.cfg.Address),
		}, true, nil
	}

	// No invalidation
	return nil, false, nil
}
