package client

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	ix_client "github.com/ethereum-optimism/optimism/indexer/client"
	"go.uber.org/zap"
)

// Config ... Client configuration
type Config struct {
	L1RpcEndpoint string
	L2RpcEndpoint string

	IndexerCfg *ix_client.Config
}

// Bundle ... Used to store all client object references
type Bundle struct {
	IxClient IxClient
	L1Client EthClient
	L2Client EthClient
	L2Geth   GethClient
}

// NewBundle ... Construct a new client bundle
func NewBundle(ctx context.Context, cfg *Config) (*Bundle, error) {
	logger := logging.WithContext(ctx)

	l1Client, err := NewEthClient(ctx, cfg.L1RpcEndpoint)
	if err != nil {
		logger.Fatal("Error creating L1 client", zap.Error(err))
		return nil, err
	}

	l2Client, err := NewEthClient(ctx, cfg.L2RpcEndpoint)
	if err != nil {
		logger.Fatal("Error creating L1 client", zap.Error(err))
		return nil, err
	}

	l2Geth, err := NewGethClient(cfg.L2RpcEndpoint)
	if err != nil {
		logger.Fatal("Error creating L2 GETH client", zap.Error(err))
		return nil, err
	}

	ixClient, err := NewIndexerClient(cfg.IndexerCfg)
	if err != nil { // Indexer client is optional so we don't want to fatal
		logger.Warn("Error creating indexer client", zap.Error(err))
	}

	return &Bundle{
		IxClient: ixClient,
		L1Client: l1Client,
		L2Client: l2Client,
		L2Geth:   l2Geth,
	}, nil
}

// FromContext ... Retrieves the client bundle from the context
func FromContext(ctx context.Context) (*Bundle, error) {
	b, err := ctx.Value(core.Clients).(*Bundle)
	if !err {
		return nil, fmt.Errorf("failed to retrieve client bundle from context")
	}

	return b, nil
}

// FromNetwork ... Retrieves an eth client from the context
func FromNetwork(ctx context.Context, n core.Network) (EthClient, error) {
	bundle, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}

	switch n {
	case core.Layer1:
		return bundle.L1Client, nil
	case core.Layer2:
		return bundle.L2Client, nil
	default:
		return nil, fmt.Errorf("invalid network supplied")
	}
}
