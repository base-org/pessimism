package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
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

// FaultDetectorCfg  ... Configuration for the fault detector invariant
type FaultDetectorCfg struct {
	L2OuptputOracle string `json:"l2_output_address"`
	L2ToL1Address   string `json:"l2_to_l1_address"`
}

// blockInfo ... Wrapper for a block
// This is used to ensure compatibility with the rollup-node software
type blockInfo struct {
	*types.Block
}

func (b blockInfo) HeaderRLP() ([]byte, error) {
	return rlp.EncodeToBytes(b.Header())
}

// blockToInfo ... Converts a block to a blockInfo
func blockToInfo(b *types.Block) blockInfo {
	return blockInfo{b}
}

// UnmarshalToFaulDetectorCfg ... Converts a general config to a fault detector invariant config
func UnmarshalToFaulDetectorCfg(isp *core.InvSessionParams) (*FaultDetectorCfg, error) {
	invConfg := FaultDetectorCfg{}
	err := json.Unmarshal(isp.Bytes(), &invConfg)
	if err != nil {
		return nil, err
	}

	return &invConfg, nil
}

// faultDetectorInv ... faultDetectorInv implementation
type faultDetectorInv struct {
	eventHash common.Hash
	cfg       *FaultDetectorCfg

	l2tol1MessagePasser  common.Address
	l2OutputOracleFilter *bindings.L2OutputOracleFilterer

	l2Client     client.EthClientInterface
	l2GethClient client.GethClient

	invariant.Invariant
}

// NewFaultDetector ... Initializer
func NewFaultDetector(ctx context.Context, cfg *FaultDetectorCfg) (invariant.Invariant, error) {
	l2Client, err := client.FromContext(ctx, core.Layer2)
	if err != nil {
		return nil, err
	}

	l1Client, err := client.FromContext(ctx, core.Layer1)
	if err != nil {
		return nil, err
	}

	l2Geth, err := client.L2GethFromContext(ctx)
	if err != nil {
		return nil, err
	}

	outputSig := crypto.Keccak256Hash([]byte("OutputProposed(bytes32,uint256,uint256,uint256)"))
	addr := common.HexToAddress(cfg.L2ToL1Address)

	outputOracle, err := bindings.NewL2OutputOracleFilterer(addr, l1Client)
	if err != nil {
		return nil, err
	}

	return &faultDetectorInv{
		cfg: cfg,

		eventHash:            outputSig,
		l2OutputOracleFilter: outputOracle,
		l2tol1MessagePasser:  addr,

		l2Client:     l2Client,
		l2GethClient: l2Geth,

		Invariant: invariant.NewBaseInvariant(core.EventLog),
	}, nil
}

// Invalidate ...
func (fd *faultDetectorInv) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation for balance invariant", zap.String("data", fmt.Sprintf("%v", td)))

	if td.Type != fd.InputType() {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	if td.Address.String() != fd.cfg.L2OuptputOracle {
		return nil, false, fmt.Errorf(invalidAddrErr, td.Address.String(), fd.cfg.L2OuptputOracle)
	}

	log, success := td.Value.(types.Log)
	if !success {
		return nil, false, fmt.Errorf(couldNotCastErr, "types.Log")
	}

	output, err := fd.l2OutputOracleFilter.ParseOutputProposed(log)
	if err != nil {
		return nil, false, err
	}

	outputBlock, err := fd.l2Client.BlockByNumber(context.Background(), output.L2BlockNumber)
	if err != nil {
		return nil, false, err
	}

	proofResp, err := fd.l2GethClient.GetProof(context.Background(),
		fd.l2tol1MessagePasser, []string{}, output.L2BlockNumber)
	if err != nil {
		return nil, false, err
	}

	asInfo := blockToInfo(outputBlock)
	expectedStateRoot, err := rollup.ComputeL2OutputRootV0(asInfo, proofResp.StorageHash)
	if err != nil {
		return nil, false, err
	}

	actualStateRoot := output.OutputRoot

	if expectedStateRoot != actualStateRoot {
		return &core.InvalOutcome{
			TimeStamp: time.Now(),
			Message:   fmt.Sprintf(faultDetectMsg, fd.cfg.L2OuptputOracle, fd.cfg.L2ToL1Address, fd.SUUID(), log.TxHash),
		}, true, nil
	}

	return nil, false, nil
}
