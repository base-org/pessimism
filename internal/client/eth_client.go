//go:generate mockgen -package mocks --destination ../mocks/eth_client.go . EthClientInterface

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
)

type EthClient struct {
	client *ethclient.Client
	restyClient *resty.Client
	url         string

}


// EthClientInterface ... Provides interface wrapper for ethClient functions
// Useful for mocking go-etheruem node client logic
type EthClientInterface interface {
	DialContext(ctx context.Context, rawURL string) error
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
}

func NewEthClient(restyClient *resty.Client) EthClientInterface {
    if restyClient == nil {
        restyClient = resty.New()
    }
	return &EthClient{
		client: &ethclient.Client{},
		restyClient: restyClient,
	}
}

func (ec *EthClient) DialContext(ctx context.Context, rawURL string) error {
	client, err := ethclient.DialContext(ctx, rawURL)

	if err != nil {
		return err
	}

	// Configure retry logic
	ec.restyClient.
		SetRetryCount(3).
		SetRetryWaitTime(5 * time.Second).
		SetRetryMaxWaitTime(20 * time.Second).
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
	var header *types.Header
	resp, err := ec.restyClient.R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"blockNumber": number.String(),
		}).
		Get(ec.url + "/eth_getBlockByNumber/{blockNumber}")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Error fetching header: %s", resp.Status())
	}

	err = json.Unmarshal(resp.Body(), &header)
	return header, err
}

func (ec *EthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	var block *types.Block
	resp, err := ec.restyClient.R().
		SetContext(ctx).
		SetPathParams(map[string]string{
			"blockNumber": number.String(),
		}).
		Get(ec.url + "/eth_getBlockByNumber/{blockNumber}")

	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Error fetching block: %s", resp.Status())
	}

	err = json.Unmarshal(resp.Body(), &block)
	return block, err
}



