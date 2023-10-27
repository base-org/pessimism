package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	p_common "github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"go.uber.org/zap"
)

const (
	greaterThanThreshold = "A withdraw was proven that is >= %f of the Optimism Portal balance"
	uncorrelatedWithdraw = "Withdrawal proven on L1 does not exist in L2ToL1MessagePasser storage"
	greaterThanPortal    = "Withdrawal amount is greater than the Optimism Portal balance"
)

const unsafeWithdrawalMsg = `
	%s
	L1PortalAddress: %s
	L2ToL1Address: %s
	
	Session UUID: %s
	L1 Proving Transaction Hash: %s
	L2 Initialization Transaction Hash: %s
	Withdrawal Size: %d
`

// UnsafeWithdrawalCfg  ... Configuration for the balance heuristic
type UnsafeWithdrawalCfg struct {
	// % of OptimismPortal balance that is considered a large withdrawal
	Threshold float64 `json:"threshold"`

	L1PortalAddress string `json:"l1_portal_address"`
	L2ToL1Address   string `json:"l2_to_l1_address"`
}

// WithdrawalSafetyHeuristic ... Withdrawal safety heuristic implementation
type WithdrawalSafetyHeuristic struct {
	ctx           context.Context
	cfg           *UnsafeWithdrawalCfg
	indexerClient client.IndexerClient

	l1Client        client.EthClient
	l1PortalFilter  *bindings.OptimismPortalFilterer
	l2ToL1MsgPasser *bindings.L2ToL1MessagePasserCaller

	heuristic.Heuristic
}

// Unmarshal ... Converts a general config to a LargeWithdrawal heuristic config
func (cfg *UnsafeWithdrawalCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &cfg)
}

// NewWithdrawalSafetyHeuristic ... Initializer
func NewWithdrawalSafetyHeuristic(ctx context.Context, cfg *UnsafeWithdrawalCfg) (heuristic.Heuristic, error) {
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
func (wi *WithdrawalSafetyHeuristic) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	// TODO - Support running from withdrawal initiated or withdrawal proven events

	// 1. Validate input
	logging.NoContext().Debug("Checking activation for withdrawal enforcement heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	if wi.ValidateInput(td) != nil {
		return nil, fmt.Errorf("invalid input supplied")
	}

	if td.Address.String() != wi.cfg.L1PortalAddress {
		return nil, fmt.Errorf(invalidAddrErr, td.Address.String(), wi.cfg.L1PortalAddress)
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, fmt.Errorf(couldNotCastErr, "types.Log")
	}

	// 2. Parse the log to a WithdrawalProven structured type
	provenWithdrawal, err := wi.l1PortalFilter.ParseWithdrawalProven(log)
	if err != nil {
		return nil, err
	}

	// 3. Get withdrawal metadata from OP Indexer API
	withdrawals, err := wi.indexerClient.GetAllWithdrawalsByAddress(provenWithdrawal.From)
	if err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, fmt.Errorf("no withdrawals found for address %s", provenWithdrawal.From.String())
	}

	// TODO - Update withdrawal decoding to convert to big.Int instead of string
	// TODO - Validate that message hash matches the proven withdrawal msg hash
	corrWithdrawal := withdrawals[0]

	// 4. Fetch the OptimismPortal balance at the time which the withdrawal was proven
	portalWEI, err := wi.l1Client.BalanceAt(context.Background(), common.HexToAddress(wi.cfg.L1PortalAddress),
		big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		return nil, err
	}

	withdrawalWEI := big.NewInt(0).SetBytes([]byte(corrWithdrawal.Amount))

	correlated, err := wi.l2ToL1MsgPasser.SentMessages(nil, provenWithdrawal.WithdrawalHash)
	if err != nil {
		return nil, err
	}

	maxAddr := common.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")
	minAddr := common.HexToAddress("0x0")

	portalETH := p_common.WeiToEther(portalWEI)
	withdrawalETH := p_common.WeiToEther(withdrawalWEI)

	// 5. Perform invariant analysis using the fetched withdrawal metadata
	invariants := []func() (bool, string){
		// 5.1
		// Check if the proven withdrawal amount is greater than the OptimismPortal value
		func() (bool, string) {
			return withdrawalWEI.Cmp(portalWEI) >= 0, greaterThanPortal
		},
		// 5.2
		// Check if the proven withdrawal amount is greater than threshold % of the OptimismPortal value
		func() (bool, string) {
			return p_common.PercentOf(withdrawalETH, portalETH).Cmp(big.NewFloat(wi.cfg.Threshold)) == 1,
				fmt.Sprintf(greaterThanThreshold, wi.cfg.Threshold)
		},
		// 5.3
		// Ensure the proven withdrawal exists in the L2ToL1MessagePasser storage
		func() (bool, string) {
			return !correlated, uncorrelatedWithdraw
		},
		// 5.4
		// Ensure message_hash != 0x0 and message_hash != 0xf...f
		func() (bool, string) {
			if corrWithdrawal.MessageHash == minAddr.String() {
				return true, "Withdrawal message hash is 0x0"
			}

			if corrWithdrawal.MessageHash == maxAddr.String() {
				return true, "Withdrawal message hash is 0xf...f"
			}

			return false, ""
		},
		// 5.5
		// Ensure that zero address and max address aren't super similar to Sorenson-Dice coefficient
		func() (bool, string) {
			c0 := p_common.SorensonDice(corrWithdrawal.MessageHash, minAddr.String())
			c1 := p_common.SorensonDice(corrWithdrawal.MessageHash, maxAddr.String())
			threshold := 0.95

			if c0 >= threshold {
				return true, "Zero address is too similar to message hash"
			}

			if c1 >= threshold {
				return true, "Max address is too similar to message hash"
			}

			return false, ""
		},
	}

	// 6. Process activation set from invariant analysis
	as := heuristic.NewActivationSet()
	for _, inv := range invariants {
		if success, msg := inv(); success {
			as = as.Add(
				&heuristic.Activation{
					TimeStamp: time.Now(),
					Message: fmt.Sprintf(unsafeWithdrawalMsg, msg,
						wi.cfg.L1PortalAddress, wi.cfg.L2ToL1Address,
						wi.SUUID(), log.TxHash.Hex(), corrWithdrawal.TransactionHash,
						withdrawalWEI),
				})
		}
	}

	return as, nil
}
