package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/common"

	"go.uber.org/zap"
)

const SupplyParityMsg = `
	%s
	Session UUID: %s
	Expected Portal Balance: %s ETH	
	Actual Portal Balance: %s ETH
	Diff Threshold: %f
`

type SupplyParityCfg struct {
	// % threshold between L1 and L2 balances
	Threshold float64 `json:"threshold"`

	L1PortalAddress string `json:"l1_portal_address"`
}

// Heuristic
type SupplyParity struct {
	ctx context.Context
	cfg *SupplyParityCfg

	heuristic.Heuristic
}

// Unmarshal ... Converts a general config to a LargeWithdrawal heuristic config
func (cfg *SupplyParityCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &cfg)
}

// NewSupplyParity ... Initializer
func NewSupplyParity(ctx context.Context, cfg *SupplyParityCfg) (heuristic.Heuristic, error) {

	return &SupplyParity{
		ctx: ctx,
		cfg: cfg,

		Heuristic: heuristic.NewBaseHeuristic(core.EventLog),
	}, nil
}

// Only runs on an arbitrarily defined event signals
func (sp *SupplyParity) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	logging.NoContext().Debug("Checking activation for supply parity heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	clients, err := client.FromContext(sp.ctx)
	if err != nil {
		return nil, err
	}

	err = sp.ValidateInput(td)
	if err != nil {
		return nil, err
	}

	ethClient, err := client.FromNetwork(sp.ctx, td.Network)
	if err != nil {
		return nil, err
	}

	assess, err := clients.IxClient.GetSupplyAssessment()
	if err != nil {
		return nil, err
	}

	expected := big.NewFloat(assess.L1DepositSum - assess.L2WithdrawalSum)
	portalAmt, err := ethClient.BalanceAt(sp.ctx, common.HexToAddress(sp.cfg.L1PortalAddress), nil)
	if err != nil {
		return nil, err
	}

	actual := new(big.Float).SetInt(portalAmt)

	// 1 - Ensure that withdrawn amount isn't greater than the deposited amount
	if assess.L2WithdrawalSum > assess.L1DepositSum {
		heuristic.NewActivationSet().
			Add(&heuristic.Activation{
				TimeStamp: time.Now(),
				Message:   fmt.Sprintf(SupplyParityMsg, "More has been withdrawn than deposited", sp.SUUID(), expected, actual, sp.cfg.Threshold),
			})

	}

	// 2 - Ensure that an amount isn't expected thats greater than the OP Portal balances
	if expected.Cmp(actual) > 0 {
		heuristic.NewActivationSet().
			Add(&heuristic.Activation{
				TimeStamp: time.Now(),
				Message:   fmt.Sprintf(SupplyParityMsg, "L1 Portal balance is less than L2 supply", sp.SUUID(), expected, actual, sp.cfg.Threshold),
			})
	}

	// 3 - Ensure that the difference between the balances is within the threshold
	diff := new(big.Float).Sub(expected, actual)
	threshold := new(big.Float).SetFloat64(sp.cfg.Threshold)
	if diff.Cmp(threshold) > 0 {
		heuristic.NewActivationSet().
			Add(&heuristic.Activation{
				TimeStamp: time.Now(),
				Message:   fmt.Sprintf(SupplyParityMsg, "L1 Portal balance is significantly greater than L2 supply", sp.SUUID(), expected, actual, sp.cfg.Threshold),
			})
	}

	return nil, nil
}
