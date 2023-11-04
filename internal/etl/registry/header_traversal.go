package registry

import (
	"context"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	ix_node "github.com/ethereum-optimism/optimism/indexer/node"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

const (
	// This could be configurable in the future
	batchSize = 100
)

type HeaderTraversal struct {
	n      core.Network
	id     core.ProcessID
	pathID core.PathID

	client       ix_node.EthClient
	traversal    *ix_node.HeaderTraversal
	pollInterval time.Duration

	stats metrics.Metricer
}

func NewHeaderTraversal(ctx context.Context, cfg *core.ClientConfig,
	opts ...process.Option) (process.Process, error) {
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

	ht := &HeaderTraversal{
		n:            cfg.Network,
		client:       node,
		traversal:    ix_node.NewHeaderTraversal(node, startHeader, big.NewInt(0)),
		pollInterval: cfg.PollInterval,
	}

	reader, err := process.NewReader(ctx, core.BlockHeader, ht, opts...)
	if err != nil {
		return nil, err
	}

	ht.id = reader.ID()
	ht.pathID = reader.PathID()
	return reader, nil
}

// Height ... Current block height
func (ht *HeaderTraversal) Height() (*big.Int, error) {
	return ht.traversal.LastHeader().Number, nil
}

func (ht *HeaderTraversal) Backfill(start, end *big.Int, consumer chan core.Event) error {
	for i := start; i.Cmp(end) < 0; i.Add(i, big.NewInt(batchSize)) {
		end := big.NewInt(0).Add(i, big.NewInt(batchSize))

		headers, err := ht.client.BlockHeadersByRange(i, end)
		if err != nil {
			return err
		}

		for _, header := range headers {
			consumer <- core.Event{
				Timestamp: time.Now(),
				Type:      core.BlockHeader,
				Value:     header,
			}
		}
	}

	return nil
}

// Loop ...
func (ht *HeaderTraversal) Loop(ctx context.Context, consumer chan core.Event) error {
	ticker := time.NewTicker(1 * time.Second)

	recent, err := ht.client.BlockHeaderByNumber(nil)
	if err != nil {
		logging.WithContext(ctx).Error("Failed to get latest header", zap.Error(err))
	}

	// backfill if provided starting header
	if ht.traversal.LastHeader() != nil {
		err = ht.Backfill(ht.traversal.LastHeader().Number, recent.Number, consumer)
		if err != nil {
			return err
		}
	} else {
		ht.traversal = ix_node.NewHeaderTraversal(ht.client, recent, big.NewInt(0))
	}

	for {
		select {
		case <-ticker.C:

			header, err := ht.client.BlockHeaderByNumber(nil)
			if err != nil {
				return err
			}

			if header.Number.Cmp(ht.traversal.LastHeader().Number) > 0 {
				headers, err := ht.traversal.NextFinalizedHeaders(batchSize)
				if err != nil {
					return err
				}

				for _, header := range headers {
					consumer <- core.Event{
						Network:   ht.n,
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
