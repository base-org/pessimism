package pipe

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	blackHoleAddress = "0x0000000000000000000000000000000000000000"
)

func extractBlackHoleTxs(td core.TransitData) ([]core.TransitData, error) {
	asBlock, success := td.Value.(types.Block)
	if !success {
		return []core.TransitData{}, fmt.Errorf("could not convert to block")
	}

	blackHoleTxs := make([]core.TransitData, 0)

	for _, tx := range asBlock.Transactions() {
		if tx.To() == nil {
			continue
		}

		if *tx.To() == common.HexToAddress(blackHoleAddress) {
			blackHoleTxs = append(blackHoleTxs, core.TransitData{
				Timestamp: td.Timestamp,
				Type:      core.BlackholeTX,
				Value:     tx,
			})
		}
	}

	return blackHoleTxs, nil
}

func NewBlackHoleTxPipe(ctx context.Context, opts ...component.Option) (component.Component, error) {
	return component.NewPipe(ctx, extractBlackHoleTxs, core.GethBlock, core.BlackholeTX, opts...)
}
