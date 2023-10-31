package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	p_common "github.com/base-org/pessimism/internal/common"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum-optimism/optimism/indexer/api/models"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
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
	Withdrawal Size: %d
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

// WithdrawalSafetyHeuristic ... Withdrawal safety heuristic implementation
type WithdrawalSafetyHeuristic struct {
	ctx           context.Context
	cfg           *WithdrawalSafetyCfg
	indexerClient client.IndexerClient

	l1Client        client.EthClient
	l1PortalFilter  *bindings.OptimismPortalFilterer
	l2ToL1MsgPasser *bindings.L2ToL1MessagePasserCaller

	heuristic.Heuristic
}

// Unmarshal ... Converts a general config to a LargeWithdrawal heuristic config
func (cfg *WithdrawalSafetyCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &cfg)
}

// NewWithdrawalSafetyHeuristic ... Initializer
func NewWithdrawalSafetyHeuristic(ctx context.Context, cfg *WithdrawalSafetyCfg) (heuristic.Heuristic, error) {
	// Validate that threshold is on exclusive range (0, 100)
	if cfg.Threshold >= 100 || cfg.Threshold < 0 {
		return nil, fmt.Errorf("invalid threshold supplied for withdrawal safety heuristic")
	}

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

	return &WithdrawalSafetyHeuristic{
		ctx: ctx,
		cfg: cfg,

		l1PortalFilter:  filter,
		l2ToL1MsgPasser: l2ToL1MsgPasser,

		indexerClient: clients.IndexerClient,
		l1Client:      clients.L1Client,

		Heuristic: heuristic.NewBaseHeuristic(core.EventLog),
	}, nil
}

// Assess ... Verifies than an L1 WithdrawalProven has a correlating hash
// to the withdrawal storage of the L2ToL1MessagePasser
// TODO - Segment this into composite functions
func (wsh *WithdrawalSafetyHeuristic) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	// TODO - Support running from withdrawal initiated or withdrawal proven events

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
	withdrawals, err := wsh.indexerClient.GetAllWithdrawalsByAddress(wm.From)
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

	parsedInt, err := strconv.Atoi(corrWithdrawal.Amount)
	if err != nil {
		return nil, err
	}

	withdrawalWEI := big.NewInt(0).SetInt64(int64(parsedInt))

	correlated, err := wsh.l2ToL1MsgPasser.SentMessages(nil, wm.Hash)
	if err != nil {
		return nil, err
	}

	invariants := wsh.GetInvariants(&corrWithdrawal, portalWEI, withdrawalWEI, correlated)

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

	msg := "\n*\t" + strings.Join(msgs, "\n*\t")
	return heuristic.NewActivationSet().Add(
		&heuristic.Activation{
			TimeStamp: time.Now(),
			Message: fmt.Sprintf(WithdrawalSafetyMsg, msg, wsh.cfg.L1PortalAddress, wsh.cfg.L2ToL1Address,
				wsh.SUUID(), log.TxHash.String(), log.BlockHash.String(), withdrawalWEI),
		},
	), nil
}

// GetInvariants ... Returns a list of invariants to be checked for in the assessment
func (wsh *WithdrawalSafetyHeuristic) GetInvariants(corrWithdrawal *models.WithdrawalItem,
	portalWEI, withdrawalWEI *big.Int, correlated bool) []func() (bool, string) {
	maxAddr := common.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	minAddr := common.HexToAddress("0x0000000000000000000000000000000000000000")

	portalAmt := new(big.Float).SetInt(portalWEI)
	withdrawAmt := new(big.Float).SetInt(withdrawalWEI)

	// Run the following invariant functions in order
	return []func() (bool, string){
		// A
		// Check if the proven withdrawal amount is greater than the OptimismPortal value
		func() (bool, string) {
			return withdrawalWEI.Cmp(portalWEI) >= 0, GreaterThanPortal
		},
		// B
		// Check if the proven withdrawal amount is greater than threshold % of the OptimismPortal value
		func() (bool, string) {
			return p_common.PercentOf(withdrawAmt, portalAmt).Cmp(big.NewFloat(wsh.cfg.Threshold)) == 1,
				fmt.Sprintf(GreaterThanThreshold, wsh.cfg.Threshold)
		},
		// C
		// Ensure the proven withdrawal exists in the L2ToL1MessagePasser storage
		func() (bool, string) {
			return !correlated, UncorrelatedWithdraw
		},
		// D
		// Ensure message_hash != 0x0...0 and message_hash != 0xf...f
		func() (bool, string) {
			if corrWithdrawal.MessageHash == minAddr.String() {
				return true, TooSimilarToZero
			}

			if corrWithdrawal.MessageHash == maxAddr.String() {
				return true, TooSimilarToMax
			}

			return false, ""
		},
		// E
		// Ensure that message isn't super similar to erroneous values using Sorenson-Dice coefficient
		func() (bool, string) {
			c0 := p_common.SorensonDice(corrWithdrawal.MessageHash, minAddr.String())
			c1 := p_common.SorensonDice(corrWithdrawal.MessageHash, maxAddr.String())
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
