package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO (#20) : Introduce optional Retry-able EthClient
// TODO (#20) : Introduce optional Retry-able EthClient
type EthClient struct {
	client *ethclient.Client
}

type EthClientInterface interface {
	DialContext(ctx context.Context, rawURL string) error
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

func (ec *EthClient) DialContext(ctx context.Context, rawURL string) error {
	client, err := ethclient.DialContext(ctx, rawURL)

	if err != nil {
		return err
	}

	ec.client = client
	return nil
}

func (ec *EthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return ec.client.HeaderByNumber(ctx, number)
}

func (ec *EthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return ec.client.BlockByNumber(ctx, number)
}
