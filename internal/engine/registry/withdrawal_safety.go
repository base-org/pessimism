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
	largeWithdrawal      = "Large withdrawal has been proven on L1"
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
func (wi *WithdrawalSafetyHeuristic) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	// TODO - Support running from withdrawal initiated or withdrawal proven events

	clients, err := client.FromContext(wi.ctx)
	if err != nil {
		return nil, err
	}

	logging.NoContext().Debug("Checking activation for withdrawal enforcement heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	// 1. Validate and extract data input
	if td.Type != wi.InputType() {
		return nil, fmt.Errorf("invalid type supplied")
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
	withdrawals, err := clients.IndexerClient.GetAllWithdrawalsByAddress(provenWithdrawal.From)
	if err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, fmt.Errorf("no withdrawals found for address %s", provenWithdrawal.From.String())
	}

	// TODO - Update withdrawal decoding to convert to big.Int instead of string
	corrWithdrawal := withdrawals[0]

	portalWEI, err := wi.l1Client.BalanceAt(context.Background(), common.HexToAddress(wi.cfg.L1PortalAddress),
		big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		return nil, err
	}

	portalETH := p_common.WeiToEther(portalWEI)

	withdrawalWEI := big.NewInt(0).SetBytes([]byte(corrWithdrawal.Amount))
	withdrawalETH := p_common.WeiToEther(withdrawalWEI)
	correlated, err := wi.l2ToL1MsgPasser.SentMessages(nil, provenWithdrawal.WithdrawalHash)
	if err != nil {
		return nil, err
	}

	// 4. Perform invariant analysis using the withdrawal metadata
	invariants := []func() (bool, string){
		// 4.1
		// Check if the proven withdrawal amount is greater than the OptimismPortal value
		func() (bool, string) {
			return withdrawalWEI.Cmp(portalWEI) >= 0, greaterThanPortal
		},
		// 4.2
		// Check if the proven withdrawal amount is greater than 5% of the OptimismPortal value
		func() (bool, string) {
			return p_common.PercentOf(withdrawalETH, portalETH).Cmp(big.NewFloat(5)) == 1, `
			A withdraw was proven that is >= 5% of the Optimism Portal balance`
		},
		// 4.3
		// Ensure the proven withdrawal exists in the L2ToL1MessagePasser storage
		func() (bool, string) {
			return !correlated, uncorrelatedWithdraw
		},
	}

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
