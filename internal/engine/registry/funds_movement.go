package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"go.uber.org/zap"
)

const fundsMovementMsg = `
	A large amount of funds were transferred to a target contract.

	Network: %s
	Target Address: %s
	Target Amount: %d
	Detected Amount: %d
	
	Session UUID: %s
	Transaction Hash: %s
`

// FundsMovementCfg  ... Configuration for the funds movement heuristic
type FundsMovementCfg struct {
	Network string `json:"network"`
	Target  string `json:"address"`
	Amount  int64  `json:"amount"`
}

// Unmarshal ... Converts a general config to a funds movement heuristic config
func (fmc *FundsMovementCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &fmc)
}

// fundMovementHeuristic ... fundMovementHeuristic implementation
type fundMovementHeuristic struct {
	cfg         *FundsMovementCfg
	n           core.Network
	transferSig string
	approveSig  string

	heuristic.Heuristic
}

// NewFundsMovement ... Initializer
func NewFundsMovement(ctx context.Context, cfg *FundsMovementCfg) (heuristic.Heuristic, error) {

	return &fundMovementHeuristic{
		n:           core.StringToNetwork(cfg.Network),
		cfg:         cfg,
		transferSig: crypto.Keccak256Hash([]byte(common.TransferEvent)).Hex(),
		approveSig:  crypto.Keccak256Hash([]byte(common.ApprovalEvent)).Hex(),

		Heuristic: heuristic.NewBaseHeuristic(core.TxReceipt),
	}, nil
}

// Assess ... Performs the funds movements heuristic logic
func (fmh *fundMovementHeuristic) Assess(td core.TransitData) ([]*core.Activation, bool, error) {
	logging.NoContext().Debug("Checking activation for funds movement heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	// 1. Validate and extract data input
	err := fmh.ValidateInput(td)
	if err != nil {
		return nil, false, err
	}

	txReceipt, ok := td.Value.(types.Receipt)

	if !ok {
		return nil, false, fmt.Errorf(couldNotCastErr, "types.Receipt")
	}

	activations := []*core.Activation{}
	// 2. Check if the transaction receipt contains an ERC-20 transfer or approval event
	for _, log := range txReceipt.Logs {
		if log.Topics[0].Hex() == fmh.transferSig ||
			log.Topics[0].Hex() == fmh.approveSig {
			// 2.a. Parse to movement event
			fundMovement, err := common.ParseMovementEvent(log.Topics)
			if err != nil {
				return nil, false, err
			}

			// 2.b. Check if the transferred amount is greater than target amount
			// and add to activations if so
			if fundMovement.Amount.Cmp(big.NewInt(fmh.cfg.Amount)) == 1 {
				activations = append(activations, &core.Activation{
					TimeStamp: time.Now(),
					Message: fmt.Sprintf(fundsMovementMsg, fmh.cfg.Network,
						fmh.cfg.Target, fmh.cfg.Amount, fundMovement.Amount,
						fmh.SUUID().String(), txReceipt.TxHash.String())},
				)
			}
		}
	}

	// 3. Return activations if any
	if len(activations) > 0 {
		return activations, true, nil
	}

	return nil, false, nil
}
