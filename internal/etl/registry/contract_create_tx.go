package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/models"
	"github.com/ethereum/go-ethereum/core/types"
)

func extractContractCreateTxs(td models.TransitData) ([]models.TransitData, error) {
	asBlock, success := td.Value.(types.Block)
	if !success {
		return []models.TransitData{}, fmt.Errorf("could not convert to block")
	}

	nilTxs := make([]models.TransitData, 0)

	for _, tx := range asBlock.Transactions() {
		if tx.To() == nil {
			nilTxs = append(nilTxs, models.TransitData{
				Timestamp: td.Timestamp,
				Type:      ContractCreateTX,
				Value:     tx,
			})
		}
	}

	return nilTxs, nil
}

func NewCreateContractTxPipe(ctx context.Context) (pipeline.Component, error) {
	return pipeline.NewPipe(ctx, extractContractCreateTxs)
}
