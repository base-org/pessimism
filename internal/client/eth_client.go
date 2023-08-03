//go:generate mockgen -package mocks --destination ../mocks/eth_client.go . EthClient

package client

/*
	NOTE
	eth client docs: https://pkg.go.dev/github.com/ethereum/go-ethereum/ethclient
	eth api docs: https://geth.ethereum.org/docs/rpc/server
*/

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"

	"github.com/base-org/pessimism/internal/core"
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

// FromContext ... Retrieves EthClient from context
func FromContext(ctx context.Context, layer core.Network) (EthClient, error) {
	key := core.L1Client
	if layer == core.Layer2 {
		key = core.L2Client
	}

	if client, ok := ctx.Value(key).(EthClient); ok {
		return client, nil
	}

	return nil, fmt.Errorf("could not load eth client object from context")
}

func GetRawL2Client(ctx context.Context) *rpc.Client {
	return ctx.Value(core.L2RawGeth).(*rpc.Client)
}

// NewEthClient ... Initializer
func NewEthClient(ctx context.Context, rawURL string) (EthClient, error) {
	return ethclient.DialContext(ctx, rawURL)
}

// RawL2EthClient ... Initializer
func RawL2EthClient(ctx context.Context, rawURL string) (*rpc.Client, error) {
	return rpc.DialContext(ctx, rawURL)

}
