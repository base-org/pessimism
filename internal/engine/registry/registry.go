package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/ethereum/go-ethereum/common"
)

// InvariantTable ... Invariant table
type InvariantTable map[core.InvariantType]*InvRegister

// InvRegister ... Invariant register struct
type InvRegister struct {
	Preprocess  func(*core.InvSessionParams) error
	Policy      core.ChainSubscription
	InputType   core.RegisterType
	Constructor func(ctx context.Context, isp *core.InvSessionParams) (invariant.Invariant, error)
}

// NewInvariantTable ... Initializer
func NewInvariantTable() InvariantTable {
	tbl := map[core.InvariantType]*InvRegister{
		core.BalanceEnforcement: {
			Preprocess:  AddressPreprocess,
			Policy:      core.BothNetworks,
			InputType:   core.AccountBalance,
			Constructor: constructBalanceEnforcement,
		},
		core.ContractEvent: {
			Preprocess:  EventPreprocess,
			Policy:      core.BothNetworks,
			InputType:   core.EventLog,
			Constructor: constructEventInv,
		},
		core.WithdrawalEnforcement: {
			Preprocess:  WithdrawEnforcePreprocess,
			Policy:      core.OnlyLayer1,
			InputType:   core.EventLog,
			Constructor: constructWithdrawalEnforce,
		},
		core.FaultDetector: {
			Preprocess:  FaultDetectPreprocess,
			Policy:      core.OnlyLayer1,
			InputType:   core.EventLog,
			Constructor: constructFaultDetector,
		},
	}

	return tbl
}

// constructEventInv ... Constructs an event invariant instance
func constructEventInv(_ context.Context, isp *core.InvSessionParams) (invariant.Invariant, error) {
	cfg := &EventInvConfig{}

	err := cfg.Unmarshal(isp)
	if err != nil {
		return nil, err
	}

	return NewEventInvariant(cfg), nil
}

// constructBalanceEnforcement ... Constructs a balance invariant instance
func constructBalanceEnforcement(_ context.Context, isp *core.InvSessionParams) (invariant.Invariant, error) {
	cfg := &BalanceInvConfig{}

	err := cfg.Unmarshal(isp)
	if err != nil {
		return nil, err
	}

	return NewBalanceInvariant(cfg)
}

// constructFaultDetector ...
func constructFaultDetector(ctx context.Context, isp *core.InvSessionParams) (invariant.Invariant, error) {
	cfg := &FaultDetectorCfg{}
	err := cfg.Unmarshal(isp)

	if err != nil {
		return nil, err
	}

	return NewFaultDetector(ctx, cfg)
}

// constructWithdrawalEnforce ...
func constructWithdrawalEnforce(ctx context.Context, isp *core.InvSessionParams) (invariant.Invariant, error) {
	cfg := &WithdrawalEnforceCfg{}
	err := cfg.Unmarshal(isp)

	if err != nil {
		return nil, err
	}

	return NewWithdrawalEnforceInv(ctx, cfg)
}

// EventPreprocess ... Ensures that an address and nested args exist in the session params
func EventPreprocess(cfg *core.InvSessionParams) error {
	err := AddressPreprocess(cfg)
	if err != nil {
		return err
	}

	if len(cfg.NestedArgs()) == 0 {
		return fmt.Errorf("no events found")
	}
	return nil
}

// AddressPreprocess ... Ensures that an address exists in the session params
func AddressPreprocess(cfg *core.InvSessionParams) error {
	nilAddr := common.Address{0}
	if cfg.Address() == nilAddr {
		return fmt.Errorf("address not found")
	}

	return nil
}

// WithdrawEnforcePreprocess ... Ensures that the l2 to l1 message passer exists
// and performs a "hack" operation to set the address key as the l2tol1MessagePasser
// address for upstream ETL components (ie. event log) to know which L1 address to
// query for events
func WithdrawEnforcePreprocess(cfg *core.InvSessionParams) error {
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
	cfg.SetValue(core.AddrKey, l1Portal)

	if len(cfg.NestedArgs()) != 0 {
		return fmt.Errorf("no nested args should be present")
	}

	cfg.SetNestedArg(WithdrawalProvenEvent)
	return nil
}

// FaultDetectPreprocess ...
func FaultDetectPreprocess(cfg *core.InvSessionParams) error {
	l2OutputOracle, err := cfg.Value(core.L2OutputOracle)
	if err != nil {
		return err
	}

	_, err = cfg.Value(core.L2ToL1MessagePasser)
	if err != nil {
		return err
	}

	cfg.SetValue(core.AddrKey, l2OutputOracle)

	if len(cfg.NestedArgs()) != 0 {
		return fmt.Errorf("no nested args should be present")
	}

	cfg.SetNestedArg(OutputProposedEvent)
	return nil
}
