package registry

import (
	"context"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/metrics"
	ix_node "github.com/ethereum-optimism/optimism/indexer/node"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	batchSize = 100

	notFoundMsg = "not found"
)

type BlockTraversal struct {
	n     core.Network
	cUUID core.CUUID
	pUUID core.PUUID

	client       ix_node.EthClient
	traversal    *ix_node.HeaderTraversal
	pollInterval time.Duration

	stats metrics.Metricer
}

func NewBlockTraversal(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	clients, err := client.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	node, err := clients.NodeClient(cfg.Network)
	if err != nil {
		return nil, err
	}

	var startHeader *types.Header
	if cfg.EndHeight != nil {
		header, err := node.BlockHeaderByNumber(cfg.EndHeight)
		if err != nil {
			return nil, err
		}

		startHeader = header
	}

	// TODO - Support network confirmation counts
	ht := ix_node.NewHeaderTraversal(node, startHeader, big.NewInt(0))

	bt := &BlockTraversal{
		n:            cfg.Network,
		client:       node,
		traversal:    ht,
		pollInterval: time.Duration(cfg.PollInterval) * time.Millisecond,
	}

	reader, err := component.NewReader(ctx, core.BlockHeader, bt, opts...)
	if err != nil {
		return nil, err
	}

	bt.cUUID = reader.UUID()
	bt.pUUID = reader.PUUID()
	return reader, nil
}

func (bt *BlockTraversal) Height() (*big.Int, error) {
	return bt.traversal.LastHeader().Number, nil
}

func (bt *BlockTraversal) Backfill(start, end *big.Int, consumer chan core.TransitData) error {
	for i := start; i.Cmp(end) < 0; i.Add(i, big.NewInt(batchSize)) {
		end := big.NewInt(0).Add(i, big.NewInt(batchSize))

		headers, err := bt.client.BlockHeadersByRange(i, end)
		if err != nil {
			return err
		}

		for _, header := range headers {
			consumer <- core.TransitData{
				Timestamp: time.Now(),
				Type:      core.BlockHeader,
				Value:     header,
			}
		}
	}

	return nil
}

// Loop ...
func (bt *BlockTraversal) Loop(ctx context.Context, consumer chan core.TransitData) error {
	ticker := time.NewTicker(1 * time.Second)

	recent, err := bt.client.BlockHeaderByNumber(nil)
	if err != nil {
		return err
	}

	// backfill if provided starting header
	if bt.traversal.LastHeader() != nil {

		bt.Backfill(bt.traversal.LastHeader().Number, recent.Number, consumer)
	} else {
		bt.traversal = ix_node.NewHeaderTraversal(bt.client, recent, big.NewInt(0))
	}

	for {
		select {
		case <-ticker.C:

			header, err := bt.client.BlockHeaderByNumber(nil)
			if err != nil {
				return err
			}

			if header.Number.Cmp(bt.traversal.LastHeader().Number) > 0 {
				headers, err := bt.traversal.NextFinalizedHeaders(batchSize)
				if err != nil {
					return err
				}

				for _, header := range headers {
					consumer <- core.TransitData{
						Network:   bt.n,
						Timestamp: time.Now(),
						Type:      core.BlockHeader,
						Value:     header,
					}
				}
			}

		case <-ctx.Done():
			return nil
		}
	}

}
