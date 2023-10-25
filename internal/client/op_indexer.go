//go:generate mockgen -package mocks --destination ../mocks/indexer_client.go . IndexerClient

package client

import (
	"github.com/ethereum-optimism/optimism/indexer/api/models"
	"github.com/ethereum-optimism/optimism/indexer/client"
	"github.com/ethereum/go-ethereum/common"
)

type IndexerClient interface {
	GetAllWithdrawalsByAddress(common.Address) ([]models.WithdrawalItem, error)
}

// NewIndexerClient ... Construct a new indexer client
func NewIndexerClient(cfg *client.Config, opts ...client.Option) (IndexerClient, error) {
	return client.NewClient(cfg, opts...)
}
