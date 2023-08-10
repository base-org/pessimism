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
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"go.uber.org/zap"
)

const largeWithdrawalMsg = `
	Large withdrawal has been proven on L1
	L1PortalAddress: %s
	L2ToL1Address: %s
	
	Session UUID: %s
	L1 Proving Transaction Hash: %s
	L2 Initialization Transaction Hash: %s
	Withdrawal Size: %d
`

// LargeWithdrawalCfg  ... Configuration for the balance heuristic
type LargeWithdrawalCfg struct {
	Threshold *big.Int `json:"threshold"`

	L1PortalAddress string `json:"l1_portal_address"`
	L2ToL1Address   string `json:"l2_to_l1_address"`
}

// LargeWithdrawHeuristic ... LargeWithdrawal heuristic implementation
type LargeWithdrawHeuristic struct {
	eventHash           common.Hash
	cfg                 *LargeWithdrawalCfg
	l2tol1MessagePasser *bindings.L2ToL1MessagePasserFilterer
	l1PortalFilter      *bindings.OptimismPortalFilterer

	heuristic.Heuristic
}

// Unmarshal ... Converts a general config to a LargeWithdrawal heuristic config
func (cfg *LargeWithdrawalCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &cfg)
}

// NewLargeWithdrawHeuristic ... Initializer
func NewLargeWithdrawHeuristic(ctx context.Context, cfg *LargeWithdrawalCfg) (heuristic.Heuristic, error) {
	l2Client, err := client.FromContext(ctx, core.Layer2)
	if err != nil {
		return nil, err
	}

	l1Client, err := client.FromContext(ctx, core.Layer1)
	if err != nil {
		return nil, err
	}

	withdrawalHash := crypto.Keccak256Hash([]byte(WithdrawalProvenEvent))

	addr := common.HexToAddress(cfg.L2ToL1Address)
	addr2 := common.HexToAddress(cfg.L1PortalAddress)
	l2MessagePasser, err := bindings.NewL2ToL1MessagePasserFilterer(addr, l2Client)
	if err != nil {
		return nil, err
	}

	filter, err := bindings.NewOptimismPortalFilterer(addr2, l1Client)
	if err != nil {
		return nil, err
	}

	return &LargeWithdrawHeuristic{
		cfg: cfg,

		eventHash:           withdrawalHash,
		l1PortalFilter:      filter,
		l2tol1MessagePasser: l2MessagePasser,

		Heuristic: heuristic.NewBaseHeuristic(core.EventLog),
	}, nil
}

// Assess ... Verifies than an L1 WithdrawalProven has a correlating hash
// to the withdrawal storage of the L2ToL1MessagePasser
func (wi *LargeWithdrawHeuristic) Assess(td core.TransitData) (*core.Activation, bool, error) {
	logging.NoContext().Debug("Checking activation for withdrawal enforcement heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	// 1. Validate and extract data input
	if td.Type != wi.InputType() {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	if td.Address.String() != wi.cfg.L1PortalAddress {
		return nil, false, fmt.Errorf(invalidAddrErr, td.Address.String(), wi.cfg.L1PortalAddress)
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, false, fmt.Errorf(couldNotCastErr, "types.Log")
	}

	// 2. Parse the log to a WithdrawalProven structured type
	provenWithdrawal, err := wi.l1PortalFilter.ParseWithdrawalProven(log)
	if err != nil {
		return nil, false, err
	}

	// 3. Check if the withdrawal exists in the message outbox of the L2ToL1MessagePasser contract
	iterator, err := wi.l2tol1MessagePasser.FilterMessagePassed(nil,
		[]*big.Int{}, []common.Address{provenWithdrawal.From}, []common.Address{provenWithdrawal.To})
	if err != nil {
		return nil, false, err
	}

	for iterator.Next() {
		if iterator.Event.WithdrawalHash == provenWithdrawal.WithdrawalHash { // Found the associated withdrawal on L2
			// 4. Check if the withdrawal amount is greater than the threshold
			if iterator.Event.Value.Cmp(wi.cfg.Threshold) == 1 {
				return &core.Activation{
					TimeStamp: time.Now(),
					Message: fmt.Sprintf(largeWithdrawalMsg,
						wi.cfg.L1PortalAddress, wi.cfg.L2ToL1Address,
						wi.SUUID(), log.TxHash.Hex(), iterator.Event.Raw.TxHash,
						iterator.Event.Value),
				}, true, nil
			}
		}
	}

	return nil, false, nil
}
