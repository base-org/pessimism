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
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"go.uber.org/zap"
)

const withdrawalEnforceMsg = `
	Proven withdrawal on L1 does not exist on L2
	L1PortalAddress: %s
	L2ToL1Address: %s
	
	Session UUID: %s
	Transaction Hash: %s
`

// WithdrawalEnforceCfg  ... Configuration for the balance heuristic
type WithdrawalEnforceCfg struct {
	L1PortalAddress string `json:"l1_portal_address"`
	L2ToL1Address   string `json:"l2_to_l1_address"`
}

// WithdrawalEnforceInv ... WithdrawalEnforceHeuristic implementation
type WithdrawalEnforceInv struct {
	eventHash           common.Hash
	cfg                 *WithdrawalEnforceCfg
	l2tol1MessagePasser *bindings.L2ToL1MessagePasserCaller
	l1PortalFilter      *bindings.OptimismPortalFilterer

	heuristic.Heuristic
}

// Unmarshal ... Converts a general config to a balance heuristic config
func (cfg *WithdrawalEnforceCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &cfg)
}

// NewWithdrawalEnforceInv ... Initializer
func NewWithdrawalEnforceInv(ctx context.Context, cfg *WithdrawalEnforceCfg) (heuristic.Heuristic, error) {
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
	l2MessagePasser, err := bindings.NewL2ToL1MessagePasserCaller(addr, l2Client)
	if err != nil {
		return nil, err
	}

	filter, err := bindings.NewOptimismPortalFilterer(addr2, l1Client)
	if err != nil {
		return nil, err
	}

	return &WithdrawalEnforceInv{
		cfg: cfg,

		eventHash:           withdrawalHash,
		l1PortalFilter:      filter,
		l2tol1MessagePasser: l2MessagePasser,

		Heuristic: heuristic.NewBaseHeuristic(core.EventLog),
	}, nil
}

// Assess ... Verifies than an L1 WithdrawalProven has a correlating hash
// to the withdrawal storage of the L2ToL1MessagePasser
func (wi *WithdrawalEnforceInv) Assess(td core.TransitData) (*core.Activation, bool, error) {
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
	exists, err := wi.l2tol1MessagePasser.SentMessages(nil, provenWithdrawal.WithdrawalHash)
	if err != nil {
		return nil, false, err
	}

	// 4. If the withdrawal does not exist, activate
	if !exists {
		return &core.Activation{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(withdrawalEnforceMsg,
				wi.cfg.L1PortalAddress, wi.cfg.L2ToL1Address, wi.SUUID(), log.TxHash.Hex()),
		}, true, nil
	}

	return nil, false, nil
}
