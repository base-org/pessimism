package registry

import (
	"context"
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

const (
	expTopicCount = 4
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
	L1PortalAddress string `json:"l1_portal"`
	L2ToL1Address   string `json:"l2_messager"`
}

// WthdrawlEnforceInv ... WithdrawalEnforceInvariant implementation
type WthdrawlEnforceInv struct {
	eventHash  common.Hash
	cfg        *WthdrawlEnforceCfg
	l2Messager *bindings.L2ToL1MessagePasserCaller

	invariant.Invariant
}

// NewWthdrawlEnforceInv ... Initializer
func NewWthdrawlEnforceInv(ctx context.Context, cfg *WthdrawlEnforceCfg) (invariant.Invariant, error) {
	l2Client, err := client.FromContext(ctx, core.Layer2)
	if err != nil {
		return nil, err
	}

	withdrawalHash := crypto.Keccak256Hash([]byte("WithdrawalProven(bytes32,address,address)"))

	addr := common.HexToAddress(cfg.L2ToL1Address)
	l2Messager, err := bindings.NewL2ToL1MessagePasserCaller(addr, l2Client)
	if err != nil {
		return nil, err
	}

	return &WthdrawlEnforceInv{
		cfg: cfg,

		eventHash:  withdrawalHash,
		l2Messager: l2Messager,

		Invariant: invariant.NewBaseInvariant(core.EventLog),
	}, nil
}

// Invalidate ... Checks if the balance is within the bounds
// specified in the config
func (wi *WthdrawlEnforceInv) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation for balance invariant", zap.String("data", fmt.Sprintf("%v", td)))

	if td.Type != wi.InputType() {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	if td.Address.String() != wi.cfg.L1PortalAddress {
		return nil, false, fmt.Errorf("invalid address supplied")
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, false, fmt.Errorf("could not convert transit data to log")
	}

	if log.Topics[0] != wi.eventHash {
		return nil, false, fmt.Errorf("invalid log topic")
	}

	if len(log.Topics) != expTopicCount {
		return nil, false, fmt.Errorf("invalid number of log topics")
	}

	withdrawalHash := log.Topics[1]
	exists, err := wi.l2Messager.SentMessages(nil, withdrawalHash)
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
