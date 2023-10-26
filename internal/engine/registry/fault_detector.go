package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
)

const faultDetectMsg = `
	Fault detection occurred
	L2OutputOracle: %s
	L2ToL1Address: %s
	
	Session UUID: %s
	Transaction Hash: %s
`

// FaultDetectorCfg  ... Configuration for the fault detector heuristic
type FaultDetectorCfg struct {
	L2OutputOracle string `json:"l2_output_address"`
	L2ToL1Address  string `json:"l2_to_l1_address"`
}

// Unmarshal ... Converts a general config to a fault detector heuristic config
func (fdc *FaultDetectorCfg) Unmarshal(isp *core.SessionParams) error {
	return json.Unmarshal(isp.Bytes(), &fdc)
}

// blockInfo ... Wrapper for a block
// This is used to ensure compatibility with the rollup-node software
type blockInfo struct {
	*types.Block
}

// HeaderRLP ... Returns the RLP encoded header of a block
func (b blockInfo) HeaderRLP() ([]byte, error) {
	return rlp.EncodeToBytes(b.Header())
}

// blockToInfo ... Converts a block to a blockInfo
func blockToInfo(b *types.Block) blockInfo {
	return blockInfo{b}
}

// faultDetectorInv ... faultDetectorInv implementation
type faultDetectorInv struct {
	eventHash common.Hash
	cfg       *FaultDetectorCfg

	l2tol1MessagePasser  common.Address
	l2OutputOracleFilter *bindings.L2OutputOracleFilterer

	l2Client     client.EthClient
	l2GethClient client.GethClient
	stats        metrics.Metricer

	heuristic.Heuristic
}

// NewFaultDetector ... Initializer
func NewFaultDetector(ctx context.Context, cfg *FaultDetectorCfg) (heuristic.Heuristic, error) {
	bundle, err := client.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	outputSig := crypto.Keccak256Hash([]byte(OutputProposedEvent))
	addr := common.HexToAddress(cfg.L2ToL1Address)

	outputOracle, err := bindings.NewL2OutputOracleFilterer(addr, bundle.L1Client)
	if err != nil {
		return nil, err
	}

	return &faultDetectorInv{
		cfg: cfg,

		eventHash:            outputSig,
		l2OutputOracleFilter: outputOracle,
		l2tol1MessagePasser:  addr,
		stats:                metrics.WithContext(ctx),

		l2Client:     bundle.L2Client,
		l2GethClient: bundle.L2Geth,

		Heuristic: heuristic.NewBaseHeuristic(core.EventLog),
	}, nil
}

// Assess ... Performs the fault detection heuristic logic
func (fd *faultDetectorInv) Assess(td core.TransitData) (*heuristic.ActivationSet, error) {
	logging.NoContext().Debug("Checking activation for fault detector heuristic",
		zap.String("data", fmt.Sprintf("%v", td)))

	// 1. Validate and extract data input
	err := fd.ValidateInput(td)
	if err != nil {
		return nil, err
	}

	if td.Address.String() != fd.cfg.L2OutputOracle {
		return nil, fmt.Errorf(invalidAddrErr, td.Address.String(), fd.cfg.L2OutputOracle)
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, fmt.Errorf(couldNotCastErr, "types.Log")
	}

	// 2. Convert raw log to structured output proposal type
	output, err := fd.l2OutputOracleFilter.ParseOutputProposed(log)
	if err != nil {
		fd.stats.RecordNodeError(core.Layer1)
		return nil, err
	}

	// 3. Fetch the L2 block with the corresponding block height of the state output
	outputBlock, err := fd.l2Client.BlockByNumber(context.Background(), output.L2BlockNumber)
	if err != nil {
		fd.stats.RecordNodeError(core.Layer2)
		return nil, err
	}

	// 4. Fetch the withdrawal state root of the L2ToL1MessagePasser contract on L2
	proofResp, err := fd.l2GethClient.GetProof(context.Background(),
		fd.l2tol1MessagePasser, []string{}, output.L2BlockNumber)
	if err != nil {
		fd.stats.RecordNodeError(core.Layer2)
		return nil, err
	}

	// 5. Compute the expected state root of the L2 block using the roll-up node software
	asInfo := blockToInfo(outputBlock)
	expectedStateRoot, err := rollup.ComputeL2OutputRootV0(asInfo, proofResp.StorageHash)
	if err != nil {
		return nil, err
	}

	actualStateRoot := output.OutputRoot

	// 6. Compare the expected state root with the actual state root; if they are not equal, then activate
	if expectedStateRoot != actualStateRoot {
		return heuristic.NewActivationSet().Add(&heuristic.Activation{
			TimeStamp: time.Now(),
			Message:   fmt.Sprintf(faultDetectMsg, fd.cfg.L2OutputOracle, fd.cfg.L2ToL1Address, fd.SUUID(), log.TxHash),
		}), nil
	}

	return heuristic.NoActivations(), nil
}
