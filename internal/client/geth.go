//go:generate mockgen -package mocks --destination ../mocks/geth_client.go . GethClient

package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// GethClient ... Provides interface wrapper for gethClient functions
type GethClient interface {
	GetProof(ctx context.Context, account common.Address, keys []string,
		blockNumber *big.Int) (*gethclient.AccountResult, error)
}

// NewGethClient ... Initializer
func NewGethClient(rawURL string) (GethClient, error) {
	rpcClient, err := rpc.Dial(rawURL)
	if err != nil {
		return nil, err
	}

	gethClient := gethclient.New(rpcClient)
	return gethClient, nil
}
