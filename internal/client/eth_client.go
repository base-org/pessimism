//go:generate mockgen -package mocks --destination ../mocks/eth_client.go . EthClientInterface

package client

/*
	NOTE
	geth client docs: https://pkg.go.dev/github.com/ethereum/go-ethereum/ethclient
	geth api docs: https://geth.ethereum.org/docs/rpc/server
*/

import (
	"context"
	"fmt"
	"math/big"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO (#20) : Introduce optional Retry-able EthClient
type EthClient struct {
	client *ethclient.Client
}

// EthClientInterface ... Provides interface wrapper for ethClient functions
// Useful for mocking go-etheruem node client logic
type EthClientInterface interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)

	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)

	BalanceAt(ctx context.Context, account common.Address, number *big.Int) (*big.Int, error)
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
}

// FromContext ... Retrieves ethClient from context
func FromContext(ctx context.Context, layer core.Network) (EthClientInterface, error) {
	key := core.L1Client
	if layer == core.Layer2 {
		key = core.L2Client
	}

	if client, ok := ctx.Value(key).(EthClientInterface); ok {
		return client, nil
	}

	return nil, fmt.Errorf("could not load eth client object from context")
}

// NewEthClient ... Initializer
func NewEthClient(ctx context.Context, rawURL string) (EthClientInterface, error) {
	client, err := ethclient.DialContext(ctx, rawURL)

	if err != nil {
		return nil, err
	}

	return &EthClient{client}, nil
}

// HeaderByNumber ... Wraps go-ethereum node headerByNumber RPC call
func (ec *EthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return ec.client.HeaderByNumber(ctx, number)
}

// BlockByNumber ... Wraps go-ethereum node blockByNumber RPC call
func (ec *EthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return ec.client.BlockByNumber(ctx, number)
}

// BalanceAt ... Wraps go-ethereum node balanceAt RPC call
func (ec *EthClient) BalanceAt(ctx context.Context, account common.Address, number *big.Int) (*big.Int, error) {
	return ec.client.BalanceAt(ctx, account, number)
}

// FilterLogs ... Wraps go-ethereum node balanceAt RPC call
func (ec *EthClient) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	return ec.client.FilterLogs(ctx, query)
}

// CallContract ... Wraps go-ethereum node callContract RPC call
func (ec *EthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return ec.client.CallContract(ctx, msg, blockNumber)
}

// CodeAt ... Wraps go-ethereum node codeAt RPC call
func (ec *EthClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	return ec.client.CodeAt(ctx, account, blockNumber)
}
