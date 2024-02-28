package registry

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/base-org/pessimism/internal/common/math"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"go.uber.org/zap"
)

// L2WithdrawalSafety ... Withdrawal safety heuristic implementation
type L2WithdrawalSafety struct {
	ctx      context.Context
	cfg      *WithdrawalSafetyCfg
	ixClient client.IxClient

	l1Client client.EthClient
	// NOTE - These values can be ingested from the chain config in the future
	l1PortalFilter  *bindings.OptimismPortalFilterer
	l2ToL1MsgPasser *bindings.L2ToL1MessagePasserCaller
	l2ToL1Filter    *bindings.L2ToL1MessagePasserFilterer

	heuristic.Heuristic
}

// NewL2WithdrawalSafety ... Initializer
func NewL2WithdrawalSafety(ctx context.Context, cfg *WithdrawalSafetyCfg) (heuristic.Heuristic, error) {
	portalAddr := common.HexToAddress(cfg.L1PortalAddress)
	l2ToL1Addr := common.HexToAddress(cfg.L2ToL1Address)

	clients, err := client.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// NOTE - All OP Stack op bindings could be moved to a single object for reuse
	// across heuristics
	filter, err := bindings.NewOptimismPortalFilterer(portalAddr, clients.L1Client)
	if err != nil {
		return nil, err
	}

	l2ToL1MsgPasser, err := bindings.NewL2ToL1MessagePasserCaller(l2ToL1Addr, clients.L2Client)
	if err != nil {
		return nil, err
	}

	l2ToL1Filter, err := bindings.NewL2ToL1MessagePasserFilterer(l2ToL1Addr, clients.L2Client)
	if err != nil {
		return nil, err
	}

	return &L2WithdrawalSafety{
		ctx: ctx,
		cfg: cfg,

		l1PortalFilter:  filter,
		l2ToL1MsgPasser: l2ToL1MsgPasser,
		l2ToL1Filter:    l2ToL1Filter,

		ixClient: clients.IxClient,
		l1Client: clients.L1Client,

		Heuristic: heuristic.New(core.Log, core.WithdrawalSafety),
	}, nil
}

// Assess ...
func (wsh *L2WithdrawalSafety) Assess(td core.Event) (*heuristic.ActivationSet, error) {
	// TODO - Support running from withdrawal finalized events as well

	// 1. Validate input
	logging.NoContext().Debug("Checking activation for withdrawal safety heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	err := wsh.Validate(td)
	if err != nil {
		return nil, err
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, fmt.Errorf(couldNotCastErr, "types.Log")
	}

	msgPassed, err := wsh.l2ToL1Filter.ParseMessagePassed(log)
	if err != nil {
		return nil, err
	}

	// 4. Fetch the OptimismPortal balance at the L1 block height which the withdrawal was proven
	portalWEI, err := wsh.l1Client.BalanceAt(context.Background(), common.HexToAddress(wsh.cfg.L1PortalAddress), nil)
	if err != nil {
		return nil, err
	}

	b := msgPassed.WithdrawalHash[0:len(msgPassed.WithdrawalHash)]

	invs := wsh.GetInvariants(portalWEI, msgPassed.Value, true)
	invs = append(invs, wsh.VerifyHash(common.BytesToHash(b))...)

	// 5. Process activation set messages from invariant analysis
	msgs := make([]string, 0)

	for _, inv := range invs {
		if success, msg := inv(); success {
			msgs = append(msgs, msg)
		}
	}

	if len(msgs) == 0 {
		return heuristic.NoActivations(), nil
	}

	msg := "\n*" + strings.Join(msgs[0:len(msgs)-1], "\n *")
	msg += msgs[len(msgs)-1]
	return heuristic.NewActivationSet().Add(
		&heuristic.Activation{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(WithdrawalSafetyMsg, msg, wsh.cfg.L1PortalAddress, wsh.cfg.L2ToL1Address,
				wsh.ID(), "N/A", log.TxHash.String(), math.WeiToEther(msgPassed.Value).String()),
		},
	), nil
}

// GetInvariants ... Returns a list of invariants to be checked for in the assessment
func (wsh *L2WithdrawalSafety) GetInvariants(portalWEI, withdrawalWEI *big.Int, correlated bool) []Invariant {
	portalAmt := new(big.Float).SetInt(portalWEI)
	withdrawAmt := new(big.Float).SetInt(withdrawalWEI)

	// Run the following invariant functions in order
	return []Invariant{
		// A
		// Check if the proven withdrawal amount is greater than the OptimismPortal value
		func() (bool, string) {
			return withdrawalWEI.Cmp(portalWEI) >= 0, GreaterThanPortal
		},
		// B
		// Check if the proven withdrawal amount is greater than threshold % of the OptimismPortal value
		func() (bool, string) {
			return math.PercentOf(withdrawAmt, portalAmt).Cmp(big.NewFloat(wsh.cfg.Threshold*100)) == 1,
				fmt.Sprintf(GreaterThanThreshold, wsh.cfg.Threshold)
		},
		// C
		// Ensure the proven withdrawal exists in the L2ToL1MessagePasser storage
		func() (bool, string) {
			return !correlated, UncorrelatedWithdraw
		},
	}
}

func (wsh *L2WithdrawalSafety) VerifyHash(hash common.Hash) []Invariant {
	maxAddr := common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	minAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")

	return []Invariant{
		// Ensure message_hash != 0x0...0 and message_hash != 0xf...f
		func() (bool, string) {
			if hash.String() == minAddr.String() {
				return true, TooSimilarToZero
			}

			if hash.String() == maxAddr.String() {
				return true, TooSimilarToMax
			}

			return false, ""
		},
		// Ensure that message isn't super similar to erroneous values using Sorenson-Dice coefficient
		func() (bool, string) {
			c0 := math.SorensonDice(hash.String(), minAddr.String())
			c1 := math.SorensonDice(hash.String(), maxAddr.String())
			threshold := wsh.cfg.CoefficientThreshold

			if c0 >= threshold {
				return true, TooSimilarToZero
			}

			if c1 >= threshold {
				return true, TooSimilarToMax
			}

			return false, ""
		},
	}
}

func (wsh *L2WithdrawalSafety) Execute(invs []Invariant, meta *WithdrawalMeta) (*heuristic.ActivationSet, error) {
	// 5. Process activation set messages from invariant analysis
	msgs := make([]string, 0)

	for _, inv := range invs {
		if success, msg := inv(); success {
			msgs = append(msgs, msg)
		}
	}

	if len(msgs) == 0 {
		return heuristic.NoActivations(), nil
	}

	msg := "\n*" + strings.Join(msgs[0:len(msgs)-1], "\n *")
	msg += msgs[len(msgs)-1]
	return heuristic.NewActivationSet().Add(
		&heuristic.Activation{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(WithdrawalSafetyMsg, msg, wsh.cfg.L1PortalAddress, wsh.cfg.L2ToL1Address,
				wsh.ID(), meta.ProvenTx.String(), meta.InitTx.String(), math.WeiToEther(meta.Value).String()),
		},
	), nil
}
