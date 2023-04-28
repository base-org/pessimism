//go:generate mockgen --destination ../../mocks/eth_client.go . EthClientInterface

package client

import (
	"context"
	"math/big"

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
	DialContext(ctx context.Context, rawURL string) error
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

// NewEthClient ... Initializer
func NewEthClient() EthClientInterface {
	return &EthClient{
		client: &ethclient.Client{},
	}
}

// DialContext ... Wraps go-etheruem node dialContext RPC creation
func (ec *EthClient) DialContext(ctx context.Context, rawURL string) error {
	client, err := ethclient.DialContext(ctx, rawURL)

	if err != nil {
		return err
	}

	ec.client = client
	return nil
}

// HeaderByNumber ... Wraps go-ethereum node headerByNumber RPC call
func (ec *EthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return ec.client.HeaderByNumber(ctx, number)
}

// BlockByNumber ... Wraps go-ethereum node blockByNumber RPC call
func (ec *EthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return ec.client.BlockByNumber(ctx, number)
}
