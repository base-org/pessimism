package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/base-org/pessimism/internal/conduit/pipeline"
	"github.com/ethereum/go-ethereum/core/types"
)

func extractContractCreateTxs(td models.TransitData) ([]models.TransitData, error) {
	asBlock, success := td.Value.(types.Block)
	if !success {
		return []models.TransitData{}, fmt.Errorf("Could not convert to block")
	}

	nilTxs := make([]models.TransitData, 0)

	for _, tx := range asBlock.Transactions() {
		if tx.To() == nil {
			nilTxs = append(nilTxs, models.TransitData{
				Timestamp: td.Timestamp,
				Type:      CONTRACT_CREATE_TX,
				Value:     tx,
			})
		}
	}

	return nilTxs, nil
}

func NewCreateContractTxPipe(ctx context.Context, inputChan chan models.TransitData) pipeline.PipelineComponent {
	return pipeline.NewPipe(ctx, extractContractCreateTxs, inputChan)
}
