package pipe

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/client"
	p_common "github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

// ReceiptDefinition ...
type ReceiptDefinition struct {
	client client.EthClient
	dlq    *p_common.DLQ[core.TransitData]

	pUUID core.PUUID
	ss    state.Store

	SK *core.StateKey
}

// NewReceiptDefinition ... Initializes the event log pipe definition
func NewReceiptDefinition(ctx context.Context, n core.Network) (*ReceiptDefinition, error) {
	// 1. Load dependencies from context
	client, err := client.FromContext(ctx, n)
	if err != nil {
		return nil, err
	}

	ss, err := state.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Construct the pipe definition
	ed := &ReceiptDefinition{
		dlq:    p_common.NewTransitDLQ(dlqMaxSize),
		client: client,
		ss:     ss,
	}
	return ed, nil
}

// NewReceiptPipe ... Initializer
func NewReceiptPipe(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	// 1. Construct the pipe definition
	ed, err := NewReceiptDefinition(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	// 2. Embed the definition into a generic pipe construction
	p, err := component.NewPipe(ctx, ed, core.GethBlock, core.TxReceipt, opts...)
	if err != nil {
		return nil, err
	}

	ed.pUUID = p.PUUID()
	return p, nil
}

// Transform ... Attempts to reprocess previously failed queries first
// before attempting to process the current block data
func (ed *ReceiptDefinition) Transform(ctx context.Context, td core.TransitData) ([]core.TransitData, error) {
	logger := logging.WithContext(ctx)

	block, ok := td.Value.(types.Block)
	if !ok {
		return nil, fmt.Errorf("invalid transit data type")
	}

	// 2. Extract receipts from the block
	txs := block.Transactions()
	receipts := make([]core.TransitData, len(txs))

	// 3. Get the receipts for each transaction
	for i, tx := range txs {
		rec, err := ed.client.TransactionReceipt(ctx, tx.Hash())
		if err != nil {
			logger.Error("failed to get transaction receipt", zap.Error(err),
				zap.String("txHash", tx.Hash().String()))
		}

		receipts[i] = core.NewTransitData(core.TxReceipt, *rec)
	}

	return receipts, nil
}
