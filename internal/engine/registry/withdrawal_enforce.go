package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
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

// WthdrawlEnforceCfg  ... Configuration for the balance invariant
type WthdrawlEnforceCfg struct {
	L1PortalAddress string `json:"l1_portal_address"`
	L2ToL1Address   string `json:"l2_to_l1_address"`
}

// UnmarshalToWthdrawlEnforceCfg ... Converts a general config to a balance invariant config
func UnmarshalToWthdrawlEnforceCfg(isp *core.InvSessionParams) (*WthdrawlEnforceCfg, error) {
	invConfg := WthdrawlEnforceCfg{}
	err := json.Unmarshal(isp.Bytes(), &invConfg)
	if err != nil {
		return nil, err
	}

	return &invConfg, nil
}

// WthdrawlEnforceInv ... WithdrawalEnforceInvariant implementation
type WthdrawlEnforceInv struct {
	eventHash           common.Hash
	cfg                 *WthdrawlEnforceCfg
	l2tol1MessagePasser *bindings.L2ToL1MessagePasserCaller
	l1PortalFilter      *bindings.OptimismPortalFilterer

	invariant.Invariant
}

// NewWthdrawlEnforceInv ... Initializer
func NewWthdrawlEnforceInv(ctx context.Context, cfg *WthdrawlEnforceCfg) (invariant.Invariant, error) {
	l2Client, err := client.FromContext(ctx, core.Layer2)
	if err != nil {
		return nil, err
	}

	l1Client, err := client.FromContext(ctx, core.Layer1)
	if err != nil {
		return nil, err
	}

	withdrawalHash := crypto.Keccak256Hash([]byte("WithdrawalProven(bytes32,address,address)"))

	addr := common.HexToAddress(cfg.L2ToL1Address)
	addr2 := common.HexToAddress(cfg.L1PortalAddress)
	l2Messager, err := bindings.NewL2ToL1MessagePasserCaller(addr, l2Client)
	if err != nil {
		return nil, err
	}

	filter, err := bindings.NewOptimismPortalFilterer(addr2, l1Client)
	if err != nil {
		return nil, err
	}

	return &WthdrawlEnforceInv{
		cfg: cfg,

		eventHash:           withdrawalHash,
		l1PortalFilter:      filter,
		l2tol1MessagePasser: l2Messager,

		Invariant: invariant.NewBaseInvariant(core.EventLog),
	}, nil
}

// Invalidate ... Verifies than an L1 WithdrawwlProven has a correlating hash
// to the withdrawal storage of the L2ToL1MessagePasser
func (wi *WthdrawlEnforceInv) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation for balance invariant", zap.String("data", fmt.Sprintf("%v", td)))

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

	provenWithdrawl, err := wi.l1PortalFilter.ParseWithdrawalProven(log)
	if err != nil {
		return nil, false, err
	}

	exists, err := wi.l2tol1MessagePasser.SentMessages(nil, provenWithdrawl.WithdrawalHash)
	if err != nil {
		return nil, false, err
	}

	if !exists { // Proven withdrawal does not exist on L1
		return &core.InvalOutcome{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(withdrawalEnforceMsg,
				wi.cfg.L1PortalAddress, wi.cfg.L2ToL1Address, wi.SUUID(), log.TxHash.Hex()),
		}, true, nil
	}

	return nil, false, nil
}
