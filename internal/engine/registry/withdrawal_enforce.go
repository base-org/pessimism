package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"go.uber.org/zap"
)

// WthdrawlEnforceCfg  ... Configuration for the balance invariant
type WthdrawlEnforceCfg struct {
	L1PortalAddress string `json:"address"`
	L2ToL1Address   string `json:"l2_messager"`
}

// WthdrawlEnforceInv ... WithdrawalEnforceInvariant implementation
type WthdrawlEnforceInv struct {
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

	addr := common.HexToAddress(cfg.L2ToL1Address)
	l2Messager, err := bindings.NewL2ToL1MessagePasserCaller(addr, l2Client)
	if err != nil {
		return nil, err
	}

	return &WthdrawlEnforceInv{
		cfg:        cfg,
		l2Messager: l2Messager,

		Invariant: invariant.NewBaseInvariant(core.EventLog, invariant.WithAddressing()),
	}, nil
}

// Invalidate ... Checks if the balance is within the bounds
// specified in the config
func (wi *WthdrawlEnforceInv) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation for balance invariant", zap.String("data", fmt.Sprintf("%v", td)))

	if td.Type != wi.InputType() {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	// WithdrawalProven(bytes32 indexed withdrawalHash, address indexed from, address indexed to)
	if td.Address.String() != wi.cfg.L1PortalAddress {
		return nil, false, fmt.Errorf("invalid address supplied")
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, false, fmt.Errorf("could not convert transit data to log")
	}

	exists, err := wi.l2Messager.SentMessages(nil, log.Topics[0])
	if err != nil {
		return nil, false, err
	}

	return nil, !exists, nil
}
