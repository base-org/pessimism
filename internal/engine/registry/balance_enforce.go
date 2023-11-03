package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/core/types"
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
	ctx    context.Context
	cfg    *BalanceInvConfig
	client client.EthClient

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
func NewBalanceHeuristic(ctx context.Context, cfg *BalanceInvConfig) (heuristic.Heuristic, error) {

	return &BalanceHeuristic{
		ctx:       ctx,
		cfg:       cfg,
		Heuristic: heuristic.NewBaseHeuristic(core.BlockHeader),
	}, nil
}

// Assess ... Checks if the balance is within the bounds
// specified in the config
func (bi *BalanceHeuristic) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	logging.NoContext().Debug("Checking activation for balance heuristic", zap.String("data", fmt.Sprintf("%v", td)))

	header, ok := td.Value.(types.Header)
	if !ok {
		return nil, fmt.Errorf(couldNotCastErr, "BlockHeader")
	}

	client, err := client.FromNetwork(bi.ctx, td.Network)
	if err != nil {
		return nil, err
	}

	// See if a tx changed the balance for the address
	balance, err := client.BalanceAt(context.Background(), td.Address, header.Number)
	if err != nil {
		return nil, err
	}

	ethBalance := float64(balance.Int64()) / 1000000000000000000

	activated := false

	// 2. Assess if balance > upper bound
	if bi.cfg.UpperBound != nil &&
		*bi.cfg.UpperBound < ethBalance {
		activated = true
	}

	// 3. Assess if balance < lower bound
	if bi.cfg.LowerBound != nil &&
		*bi.cfg.LowerBound > ethBalance {
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

		msg := fmt.Sprintf(reportMsg, balance, upper, lower, bi.SUUID(), bi.cfg.Address)

		return heuristic.NewActivationSet().Add(
			&heuristic.Activation{
				Message:   msg,
				TimeStamp: time.Now(),
			}), nil
	}

	// No activation
	return heuristic.NoActivations(), nil
}
