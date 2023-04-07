package registry

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func extractBlackHoleTxs(td models.TransitData) ([]models.TransitData, error) {
	asBlock, success := td.Value.(types.Block)
	if !success {
		return []models.TransitData{}, fmt.Errorf("could not convert to block")
	}

	blackHoleTxs := make([]models.TransitData, 0)

	for _, tx := range asBlock.Transactions() {
		if tx.To() == nil {
			continue
		}

		if *tx.To() == common.HexToAddress("0x0") {
			blackHoleTxs = append(blackHoleTxs, models.TransitData{
				Timestamp: td.Timestamp,
				Type:      ContractCreateTX,
				Value:     tx,
			})
		}
	}

	return blackHoleTxs, nil
}

func NewBlackHoleTxPipe(ctx context.Context, opts ...component.Option) (component.Component, error) {
	return component.NewPipe(ctx, extractContractCreateTxs, BlackholeTX, opts...)
}
