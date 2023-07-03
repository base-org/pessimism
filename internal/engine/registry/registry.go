package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
)

// InvariantTable ... Invariant table
type InvariantTable map[core.InvariantType]*InvRegister

// InvRegister ... Invariant register struct
type InvRegister struct {
	Preprocess  func(*core.InvSessionParams) error
	MultiChain  bool
	InputType   core.RegisterType
	Constructor func(ctx context.Context, isp *core.InvSessionParams) (invariant.Invariant, error)
}

// NewInvariantTable ... Initializer
func NewInvariantTable() InvariantTable {
	tbl := map[core.InvariantType]*InvRegister{
		core.BalanceEnforcement: {
			Preprocess:  addressPreprocess,
			MultiChain:  false,
			InputType:   core.AccountBalance,
			Constructor: constructBalanceInv,
		},
		core.ContractEvent: {
			Preprocess:  eventPreprocess,
			MultiChain:  false,
			InputType:   core.EventLog,
			Constructor: constructEventInv,
		},
		core.WithdrawalEnforcement: {
			Preprocess:  preprocWithdrwlEnforce,
			MultiChain:  true,
			InputType:   core.EventLog,
			Constructor: constructWithdrawlEnforceInv,
		},
	}

	return tbl
}

func constructWithdrawlEnforceInv(ctx context.Context, isp *core.InvSessionParams) (invariant.Invariant, error) {
	cfg, err := UnmarshalToWthdrawlEnforceCfg(isp)
	if err != nil {
		return nil, err
	}

	return NewWthdrawlEnforceInv(ctx, cfg)
}

func constructEventInv(_ context.Context, isp *core.InvSessionParams) (invariant.Invariant, error) {
	cfg, err := UnmarshalToEventInvConfig(isp)
	if err != nil {
		return nil, err
	}

	return NewEventInvariant(cfg)
}

func constructBalanceInv(_ context.Context, isp *core.InvSessionParams) (invariant.Invariant, error) {
	cfg, err := UnmarshalToBalanceInvConfig(isp)
	if err != nil {
		return nil, err
	}

	return NewBalanceInvariant(cfg)
}

// eventPreprocess ... Ensures that an address and nesteed args exist in the session params
func eventPreprocess(cfg *core.InvSessionParams) error {
	err := addressPreprocess(cfg)
	if err != nil {
		return err
	}

	if len(cfg.NestedArgs()) == 0 {
		return fmt.Errorf("no events found")
	}
	return nil
}

// NewBalanceInvariant ... Ensures that an address exists in the session params
func addressPreprocess(cfg *core.InvSessionParams) error {
	if cfg.Address() == "" {
		return fmt.Errorf("address not found")
	}

	return nil
}

// preprocWithdrwlEnforce ... Ensures that the l2 to l1 message passer exists
// and performs a "hack" operation to set the address key as the l2tol1MessagePasser
// address for upstream ETL components (ie. event log) to know which address to
// query for events
func preprocWithdrwlEnforce(cfg *core.InvSessionParams) error {
	l1Portal, err := cfg.Value(core.L1Portal)
	if err != nil {
		return err
	}

	// Configure the session to subscribe to events from the L1Portal contract
	cfg.SetValue(core.AddrKey, l1Portal)

	if len(cfg.NestedArgs()) != 0 {
		return fmt.Errorf("No nested args should be present")
	}

	cfg.SetNestedArg("WithdrawalProven(bytes32,address,address)")
	return nil
}
