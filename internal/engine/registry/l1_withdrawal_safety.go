package registry

import (
	"context"
	"encoding/json"
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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"go.uber.org/zap"
)

const (
	GreaterThanThreshold = "A withdraw was proven that is >= %f of the Optimism Portal balance"
	UncorrelatedWithdraw = "Withdrawal proven on L1 does not exist in L2ToL1MessagePasser storage"
	GreaterThanPortal    = "Withdrawal amount is greater than the Optimism Portal balance"
	TooSimilarToZero     = "Withdrawal message hash is too similar to zero address"
	TooSimilarToMax      = "Withdrawal message hash is too similar to max address"
)

const WithdrawalSafetyMsg = `
	%s
	L1PortalAddress: %s
	L2ToL1Address: %s
	
	Session UUID: %s
	L1 Proving Transaction Hash: %s
	L2 Initialization Transaction Hash: %s
	Withdrawal Size: %s ETH
`

type WithdrawMetadata struct {
	Hash common.Hash
	From common.Address
	To   common.Address
}

func MetaFromProven(log types.Log, filter *bindings.OptimismPortalFilterer) (*WithdrawMetadata, error) {
	provenWithdrawal, err := filter.ParseWithdrawalProven(log)
	if err != nil {
		return nil, err
	}

	return &WithdrawMetadata{
		Hash: provenWithdrawal.WithdrawalHash,
		From: provenWithdrawal.From,
		To:   provenWithdrawal.To,
	}, nil
}

// WithdrawalSafetyCfg  ... Configuration for the balance heuristic
type WithdrawalSafetyCfg struct {
	// % of OptimismPortal balance that is considered a large withdrawal
	Threshold float64 `json:"threshold"`

	// Sorenson-Dice coefficient threshold for zero address and max address
	CoefficientThreshold float64 `json:"coefficient_threshold"`

	L1PortalAddress string `json:"l1_portal_address"`
	L2ToL1Address   string `json:"l2_to_l1_address"`
}

// L1WithdrawalSafety ... Withdrawal safety heuristic implementation
type L1WithdrawalSafety struct {
	ctx      context.Context
	cfg      *WithdrawalSafetyCfg
	ixClient client.IxClient

	l1Client client.EthClient
	// NOTE - These values can be ingested from the chain config in the future
	l1PortalFilter  *bindings.OptimismPortalFilterer
	l2ToL1MsgPasser *bindings.L2ToL1MessagePasserCaller

	*L2WithdrawalSafety
}

// Unmarshal ... Converts a general config to a LargeWithdrawal heuristic config
func (cfg *WithdrawalSafetyCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &cfg)
}

// NewL1WithdrawalSafety ... Initializer
func NewL1WithdrawalSafety(ctx context.Context, cfg *WithdrawalSafetyCfg) (heuristic.Heuristic, error) {
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

	wsh, err := NewL2WithdrawalSafety(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &L1WithdrawalSafety{
		ctx: ctx,
		cfg: cfg,

		l1PortalFilter:  filter,
		l2ToL1MsgPasser: l2ToL1MsgPasser,

		ixClient: clients.IxClient,
		l1Client: clients.L1Client,

		L2WithdrawalSafety: wsh.(*L2WithdrawalSafety),
	}, nil
}

// Assess ...
func (wsh *L1WithdrawalSafety) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	// TODO - Support running from withdrawal finalized events as well

	// 1. Validate input
	logging.NoContext().Debug("Checking activation for withdrawal enforcement heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	err := wsh.ValidateInput(td)
	if err != nil {
		return nil, err
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, fmt.Errorf(couldNotCastErr, "types.Log")
	}

	var wm *WithdrawMetadata
	switch log.Topics[0] {
	// TODO(#178) - Feat - Support WithdrawalProven processing in withdrawal_safety heuristic
	// case WithdrawalFinalSig:
	// 	wm, err = MetaFromFinalized(log, wi.l1PortalFilter)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	case WithdrawalProvenSig:
		wm, err = MetaFromProven(log, wsh.l1PortalFilter)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("invalid topic supplied")
	}

	// 3. Get withdrawal metadata from OP Indexer API
	withdrawals, err := wsh.ixClient.GetAllWithdrawalsByAddress(wm.From)
	if err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, fmt.Errorf("no withdrawals found for address %s", wm.From.String())
	}

	// TODO - Update withdrawal decoding in client to convert to big.Int instead of string
	corrWithdrawal := withdrawals[0]

	// TODO - Validate that message hash matches the proven withdrawal msg hash
	// if corrWithdrawal.TransactionHash != wm.Hash.String() {
	// 	return nil, fmt.Errorf("withdrawal hash mismatch, expected %s, got %s", wm.Hash.String(),
	// 		corrWithdrawal.TransactionHash)
	// }

	// 4. Fetch the OptimismPortal balance at the L1 block height which the withdrawal was proven
	portalWEI, err := wsh.l1Client.BalanceAt(context.Background(), common.HexToAddress(wsh.cfg.L1PortalAddress),
		big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		return nil, err
	}

	b := []byte(corrWithdrawal.Amount)
	withdrawalWEI := big.NewInt(0).SetBytes(b)

	correlated, err := wsh.l2ToL1MsgPasser.SentMessages(&bind.CallOpts{
		BlockNumber: big.NewInt(int64(log.BlockNumber)),
	}, wm.Hash)
	if err != nil {
		return nil, err
	}

	h := common.HexToHash(corrWithdrawal.TransactionHash)
	invariants := wsh.GetInvariants(h, portalWEI, withdrawalWEI, correlated)

	// 5. Process activation set messages from invariant analysis
	msgs := make([]string, 0)

	for _, inv := range invariants {
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
				wsh.SUUID(), log.TxHash.String(), corrWithdrawal.TransactionHash, math.WeiToEther(withdrawalWEI).String()),
		},
	), nil
}
