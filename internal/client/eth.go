//go:generate mockgen -package mocks --destination ../mocks/eth_client.go . EthClient

package client

/*
	NOTE
	eth client docs: https://pkg.go.dev/github.com/ethereum/go-ethereum/ethclient
	eth api docs: https://geth.ethereum.org/docs/rpc/server
*/

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO (#20) : Introduce optional Retry-able EthClient

// EthClient ... Provides interface wrapper for ethClient functions
// Useful for mocking go-ethereum json rpc client logic
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

// NewEthClient ... Initializer
func NewEthClient(ctx context.Context, rawURL string) (EthClient, error) {
	return ethclient.DialContext(ctx, rawURL)
}
