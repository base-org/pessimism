//go:generate mockgen -package mocks --destination ../mocks/eth_client.go . EthClientInterface

package client

import (
	"context"
	"math/big"
	"time"

	"github.com/avast/retry-go"
	"github.com/base-org/pessimism/internal/config"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
)

type EthClient struct {
	client           *ethclient.Client
	restyClient      *resty.Client
	url              string
	retryCount       int
	retryWaitTime    time.Duration
	retryMaxWaitTime time.Duration
}

// EthClientInterface ... Provides interface wrapper for ethClient functions
// Useful for mocking go-etheruem node client logic
type EthClientInterface interface {
	DialContext(ctx context.Context, rawURL string) error
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

func NewEthClient(restyClient *resty.Client, ethClientCfg *config.EthClientCfg) EthClientInterface {
	if restyClient == nil {
		restyClient = resty.New()
	}

	restyDebug := ethClientCfg.RetryConfig.Debug
	restyClient.SetDebug(restyDebug)

	return &EthClient{
		client:           &ethclient.Client{},
		restyClient:      restyClient,
		retryCount:       ethClientCfg.RetryConfig.RetryCount,
		retryWaitTime:    time.Duration(ethClientCfg.RetryConfig.RetryWaitTime) * time.Second,
		retryMaxWaitTime: time.Duration(ethClientCfg.RetryConfig.RetryMaxWaitTime) * time.Second,
	}
}

func (ec *EthClient) DialContext(ctx context.Context, rawURL string) error {
	client, err := ethclient.DialContext(ctx, rawURL)

	if err != nil {
		return err
	}

	ec.restyClient.
		SetRetryCount(ec.retryCount).
		SetRetryWaitTime(ec.retryWaitTime).
		SetRetryMaxWaitTime(ec.retryMaxWaitTime).
		AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return r.IsError() || err != nil
			},
		)

	ec.client = client
	ec.url = rawURL // Store the URL
	return nil
}

func (ec *EthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	retryableFunc := func() (*types.Header, error) {
		header, err := ec.client.HeaderByNumber(ctx, number)
		if err != nil {
			return nil, err
		}
		return header, nil
	}

	// Execute the function with retry logic
	var header *types.Header
	err := retry.Do(
		func() error {
			var err error
			header, err = retryableFunc()
			return err
		},
		retry.Attempts(uint(ec.retryCount)),
		retry.Delay(ec.retryWaitTime),
		retry.MaxDelay(ec.retryMaxWaitTime),
	)

	if err != nil {
		return nil, err
	}
	return header, nil
}

func (ec *EthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	retryableFunc := func() (*types.Block, error) {
		block, err := ec.client.BlockByNumber(ctx, number)
		if err != nil {
			return nil, err
		}
		return block, nil
	}

	// Execute the function with retry logic
	var block *types.Block
	err := retry.Do(
		func() error {
			var err error
			block, err = retryableFunc()
			return err
		},
		retry.Attempts(uint(ec.retryCount)),
		retry.Delay(ec.retryWaitTime),
		retry.MaxDelay(ec.retryMaxWaitTime),
	)

	if err != nil {
		return nil, err
	}
	return block, nil
}
