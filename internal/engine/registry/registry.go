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
type HeuristicTable map[core.HeuristicType]*Registry

type Registry struct {
	PrepareValidate func(*core.SessionParams) error
	Policy          core.ChainSubscription
	InputType       core.TopicType
	Constructor     func(ctx context.Context, isp *core.SessionParams) (heuristic.Heuristic, error)
}

type Invariant func() (bool, string)

// NewHeuristicTable ... Initializer
func NewHeuristicTable() HeuristicTable {
	tbl := map[core.HeuristicType]*Registry{
		core.BalanceEnforcement: {
			PrepareValidate: ValidateAddressing,
			Policy:          core.BothNetworks,
			InputType:       core.BlockHeader,
			Constructor:     constructBalanceEnforcement,
		},
		core.ContractEvent: {
			PrepareValidate: ValidateTracking,
			Policy:          core.BothNetworks,
			InputType:       core.Log,
			Constructor:     constructEventInv,
		},
		core.FaultDetector: {
			PrepareValidate: FaultDetectionPrepare,
			Policy:          core.OnlyLayer1,
			InputType:       core.Log,
			Constructor:     constructFaultDetector,
		},
		core.WithdrawalSafety: {
			PrepareValidate: WithdrawHeuristicPrep,
			Policy:          core.OnlyLayer1,
			InputType:       core.Log,
			Constructor:     constructWithdrawalSafety,
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
func constructBalanceEnforcement(ctx context.Context, isp *core.SessionParams) (heuristic.Heuristic, error) {
	cfg := &BalanceInvConfig{}

	err := cfg.Unmarshal(isp)
	if err != nil {
		return nil, err
	}

	return NewBalanceHeuristic(ctx, cfg)
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

// constructWithdrawalSafety ... Constructs a withdrawal safety heuristic instance
func constructWithdrawalSafety(ctx context.Context, isp *core.SessionParams) (heuristic.Heuristic, error) {
	cfg := &WithdrawalSafetyCfg{}
	err := cfg.Unmarshal(isp)

	if err != nil {
		return nil, err
	}

	// Validate that thresholds are on exclusive range (0, 1)
	if cfg.Threshold >= 1 || cfg.Threshold <= 0 {
		return nil, fmt.Errorf("invalid threshold supplied for withdrawal safety heuristic")
	}

	if cfg.CoefficientThreshold >= 1 || cfg.CoefficientThreshold <= 0 {
		return nil, fmt.Errorf("invalid coefficient threshold supplied for withdrawal safety heuristic")
	}

	switch isp.Net {
	case core.Layer1:
		return NewL1WithdrawalSafety(ctx, cfg)

	case core.Layer2:
		return NewL2WithdrawalSafety(ctx, cfg)

	default:
		return nil, fmt.Errorf("invalid network supplied for withdrawal safety heuristic")
	}
}

// ValidateTracking ... Ensures that an address and nested args exist in the session params
func ValidateTracking(cfg *core.SessionParams) error {
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

// WithdrawHeuristicPrep ... Ensures that the l2 to l1 message passer exists
// and performs a "hack" operation to set the address key as the l2tol1MessagePasser
// address for upstream ETL process (ie. event log) to know which L1 address to
// query for events
func WithdrawHeuristicPrep(cfg *core.SessionParams) error {
	l1Portal, err := cfg.Value(core.L1Portal)
	if err != nil {
		return err
	}

	l2MsgPasser, err := cfg.Value(core.L2ToL1MessagePasser)
	if err != nil {
		return err
	}

	err = ValidateNoTopicsExist(cfg)
	if err != nil {
		return err
	}

	switch cfg.Net {
	case core.Layer1:
		cfg.SetValue(logging.AddrKey, l1Portal)
		cfg.SetNestedArg(WithdrawalProvenEvent)
		// cfg.SetNestedArg(WithdrawalFinalEvent)
	case core.Layer2:
		cfg.SetValue(logging.AddrKey, l2MsgPasser)
		cfg.SetNestedArg(MessagePassed)
	}

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
