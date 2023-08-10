package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/common"
)

// HeuristicTable ... Heuristic table
type HeuristicTable map[core.HeuristicType]*InvRegister

// InvRegister ... Heuristic register struct
type InvRegister struct {
	PrepareValidate func(*core.SessionParams) error
	Policy          core.ChainSubscription
	InputType       core.RegisterType
	Constructor     func(ctx context.Context, isp *core.SessionParams) (heuristic.Heuristic, error)
}

// NewHeuristicTable ... Initializer
func NewHeuristicTable() HeuristicTable {
	tbl := map[core.HeuristicType]*InvRegister{
		core.BalanceEnforcement: {
			PrepareValidate: ValidateAddressing,
			Policy:          core.BothNetworks,
			InputType:       core.AccountBalance,
			Constructor:     constructBalanceEnforcement,
		},
		core.ContractEvent: {
			PrepareValidate: ValidateEventTracking,
			Policy:          core.BothNetworks,
			InputType:       core.EventLog,
			Constructor:     constructEventInv,
		},
		core.FaultDetector: {
			PrepareValidate: FaultDetectionPrepare,
			Policy:          core.OnlyLayer1,
			InputType:       core.EventLog,
			Constructor:     constructFaultDetector,
		},
		core.WithdrawalEnforcement: {
			PrepareValidate: WithdrawEnforcePrepare,
			Policy:          core.OnlyLayer1,
			InputType:       core.EventLog,
			Constructor:     constructWithdrawalEnforce,
		},
		core.LargeWithdrawal: {
			PrepareValidate: WithdrawEnforcePrepare,
			Policy:          core.OnlyLayer1,
			InputType:       core.EventLog,
			Constructor:     constructLargeWithdrawal,
		},
	}

	return tbl
}

// constructEventInv ... Constructs an event heuristic instance
func constructEventInv(_ context.Context, isp *core.SessionParams) (heuristic.Heuristic, error) {
	cfg := &EventInvConfig{}

	err := cfg.Unmarshal(isp)
	if err != nil {
		return nil, err
	}

	return NewEventHeuristic(cfg), nil
}

// constructBalanceEnforcement ... Constructs a balance heuristic instance
func constructBalanceEnforcement(_ context.Context, isp *core.SessionParams) (heuristic.Heuristic, error) {
	cfg := &BalanceInvConfig{}

	err := cfg.Unmarshal(isp)
	if err != nil {
		return nil, err
	}

	return NewBalanceHeuristic(cfg)
}

// constructFaultDetector ... Constructs a fault detector heuristic instance
func constructFaultDetector(ctx context.Context, isp *core.SessionParams) (heuristic.Heuristic, error) {
	cfg := &FaultDetectorCfg{}
	err := cfg.Unmarshal(isp)

	if err != nil {
		return nil, err
	}

	return NewFaultDetector(ctx, cfg)
}

// constructWithdrawalEnforce ... Constructs a withdrawal enforcement heuristic instance
func constructWithdrawalEnforce(ctx context.Context, isp *core.SessionParams) (heuristic.Heuristic, error) {
	cfg := &WithdrawalEnforceCfg{}
	err := cfg.Unmarshal(isp)

	if err != nil {
		return nil, err
	}

	return NewWithdrawalEnforceInv(ctx, cfg)
}

// constructLargeWithdrawal ... Constructs a large withdrawal heuristic instance
func constructLargeWithdrawal(ctx context.Context, isp *core.SessionParams) (heuristic.Heuristic, error) {
	cfg := &LargeWithdrawalCfg{}
	err := cfg.Unmarshal(isp)

	if err != nil {
		return nil, err
	}

	return NewLargeWithdrawHeuristic(ctx, cfg)
}

// ValidateEventTracking ... Ensures that an address and nested args exist in the session params
func ValidateEventTracking(cfg *core.SessionParams) error {
	err := ValidateAddressing(cfg)
	if err != nil {
		return err
	}

	return ValidateTopicsExist(cfg)
}

// ValidateAddressing ... Ensures that an address exists in the session params
func ValidateAddressing(cfg *core.SessionParams) error {
	nilAddr := common.Address{0}
	if cfg.Address() == nilAddr {
		return fmt.Errorf(zeroAddressErr)
	}

	return nil
}

// ValidateTopicsExist ... Ensures that some nested args exist in the session params
func ValidateTopicsExist(cfg *core.SessionParams) error {
	if len(cfg.NestedArgs()) == 0 {
		return fmt.Errorf(noNestedArgsErr)
	}
	return nil
}

// ValidateNoTopicsExist ... Ensures that no nested args exist in the session params
func ValidateNoTopicsExist(cfg *core.SessionParams) error {
	if len(cfg.NestedArgs()) != 0 {
		return fmt.Errorf(noNestedArgsErr)
	}
	return nil
}

// WithdrawEnforcePrepare ... Ensures that the l2 to l1 message passer exists
// and performs a "hack" operation to set the address key as the l2tol1MessagePasser
// address for upstream ETL components (ie. event log) to know which L1 address to
// query for events
func WithdrawEnforcePrepare(cfg *core.SessionParams) error {
	l1Portal, err := cfg.Value(core.L1Portal)
	if err != nil {
		return err
	}

	_, err = cfg.Value(core.L2ToL1MessagePasser)
	if err != nil {
		return err
	}

	// Configure the session to inform the ETL to subscribe
	// to withdrawal proof events from the L1Portal contract
	cfg.SetValue(logging.AddrKey, l1Portal)

	err = ValidateNoTopicsExist(cfg)
	if err != nil {
		return err
	}

	cfg.SetNestedArg(WithdrawalProvenEvent)
	return nil
}

// FaultDetectionPrepare ... Configures the session params with the appropriate
// address key and nested args for the ETL to subscribe to L2OutputOracle events
func FaultDetectionPrepare(cfg *core.SessionParams) error {
	l2OutputOracle, err := cfg.Value(core.L2OutputOracle)
	if err != nil {
		return err
	}

	_, err = cfg.Value(core.L2ToL1MessagePasser)
	if err != nil {
		return err
	}

	err = ValidateNoTopicsExist(cfg)
	if err != nil {
		return err
	}

	cfg.SetValue(logging.AddrKey, l2OutputOracle)

	cfg.SetNestedArg(OutputProposedEvent)
	return nil
}
