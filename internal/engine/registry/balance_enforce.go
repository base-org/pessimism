package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// BalanceInvConfig ... Configuration for the balance heuristic
type BalanceInvConfig struct {
	Address    string   `json:"address"`
	UpperBound *float64 `json:"upper"`
	LowerBound *float64 `json:"lower"`
}

// Unmarshal ... Converts a general config to a balance heuristic config
func (bi *BalanceInvConfig) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &bi)
}

// BalanceHeuristic ...
type BalanceHeuristic struct {
	cfg *BalanceInvConfig

	heuristic.Heuristic
}

// reportMsg ... Message to be sent to the alerting subsystem
const reportMsg = `
	Current value: %3f
	Upper bound: %s
	Lower bound: %s

	Session UUID: %s
	Session Address: %s 
`

// NewBalanceHeuristic ... Initializer
func NewBalanceHeuristic(cfg *BalanceInvConfig) (heuristic.Heuristic, error) {
	return &BalanceHeuristic{
		cfg:       cfg,
		Heuristic: heuristic.NewBaseHeuristic(core.AccountBalance),
	}, nil
}

// Assess ... Checks if the balance is within the bounds
// specified in the config
func (bi *BalanceHeuristic) Assess(td core.TransitData) ([]*core.Activation, bool, error) {
	logging.NoContext().Debug("Checking activation for balance heuristic", zap.String("data", fmt.Sprintf("%v", td)))

	// 1. Validate and extract balance input
	err := bi.ValidateInput(td)
	if err != nil {
		return nil, false, err
	}

	balance, ok := td.Value.(float64)
	if !ok {
		return nil, false, fmt.Errorf(couldNotCastErr, "float64")
	}

	activated := false

	// 2. Assess if balance > upper bound
	if bi.cfg.UpperBound != nil &&
		*bi.cfg.UpperBound < balance {
		activated = true
	}

	// 3. Assess if balance < lower bound
	if bi.cfg.LowerBound != nil &&
		*bi.cfg.LowerBound > balance {
		activated = true
	}

	/// 4. Generate activation outcome if activated
	if activated {
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

		return []*core.Activation{{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(reportMsg, balance,
				upper, lower,
				bi.SUUID(), bi.cfg.Address),
		}}, true, nil
	}

	// No activation
	return nil, false, nil
}
