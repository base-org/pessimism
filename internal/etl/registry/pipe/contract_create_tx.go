package pipe

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/ethereum/go-ethereum/core/types"
)

func extractContractCreateTxs(td core.TransitData) ([]core.TransitData, error) {
	asBlock, success := td.Value.(types.Block)
	if !success {
		return []core.TransitData{}, fmt.Errorf("could not convert to block")
	}

	nilTxs := make([]core.TransitData, 0)

	for _, tx := range asBlock.Transactions() {
		if tx.To() == nil {
			nilTxs = append(nilTxs, core.TransitData{
				Timestamp: td.Timestamp,
				Type:      core.ContractCreateTX,
				Value:     tx,
			})
		}
	}

	return nilTxs, nil
}

func NewCreateContractTxPipe(ctx context.Context, opts ...component.Option) (component.Component, error) {
	return component.NewPipe(ctx, extractContractCreateTxs, core.GethBlock, core.ContractCreateTX, opts...)
}
