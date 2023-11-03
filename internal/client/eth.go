//go:generate mockgen -package mocks --destination ../mocks/eth_client.go . EthClient,NodeClient

package client

import (
	"context"
	"math/big"

	"github.com/base-org/pessimism/internal/metrics"
	ix_node "github.com/ethereum-optimism/optimism/indexer/node"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)

	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)

	BalanceAt(ctx context.Context, account common.Address, number *big.Int) (*big.Int, error)
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery,
		ch chan<- types.Log) (ethereum.Subscription, error)
}

type NodeClient interface {
	BlockHeaderByNumber(*big.Int) (*types.Header, error)
	BlockHeaderByHash(common.Hash) (*types.Header, error)
	BlockHeadersByRange(*big.Int, *big.Int) ([]types.Header, error)

	TxByHash(common.Hash) (*types.Transaction, error)

	StorageHash(common.Address, *big.Int) (common.Hash, error)
	FilterLogs(ethereum.FilterQuery) ([]types.Log, error)
}

// NewEthClient ... Initializer
func NewEthClient(ctx context.Context, rawURL string) (EthClient, error) {
	return ethclient.DialContext(ctx, rawURL)
}

func NewNodeClient(ctx context.Context, rpcURL string) (NodeClient, error) {
	stats := metrics.WithContext(ctx)

	return ix_node.DialEthClient(rpcURL, stats)
}
