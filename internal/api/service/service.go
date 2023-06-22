//go:generate mockgen -package mocks --destination ../../mocks/api_service.go . Service

package service

import (
	"context"
	"fmt"
	"github.com/base-org/pessimism/internal/client"
	"time"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/subsystem"
)

// Config ... Used to store necessary API service config values
type Config struct {
	L1RpcEndpoint  string
	L2RpcEndpoint  string
	L1PollInterval int
	L2PollInterval int
	SlackURL       string
}

// GetEndpointForNetwork ... Returns config endpoint for network type
func (cfg *Config) GetEndpointForNetwork(n core.Network) (string, error) {
	switch n {
	case core.Layer1:
		return cfg.L1RpcEndpoint, nil

	case core.Layer2:
		return cfg.L2RpcEndpoint, nil

	default:
		return "", fmt.Errorf("could not find endpoint for network %s", n.String())
	}
}

// GetEndpointForNetwork ... Returns config poll-interval for network type
func (cfg *Config) GetPollIntervalForNetwork(n core.Network) (time.Duration, error) {
	switch n {
	case core.Layer1:
		return time.Duration(cfg.L1PollInterval), nil

	case core.Layer2:
		return time.Duration(cfg.L2PollInterval), nil

	default:
		return 0, fmt.Errorf("could not find endpoint for network %s", n.String())
	}
}

// Service ... Interface for API service
type Service interface {
	ProcessInvariantRequest(ir models.InvRequestBody) (core.InvSessionUUID, error)
	CheckHealth() *models.HealthCheck
	CheckETHRPCHealth(url string) bool
}

// PessimismService ... API service
type PessimismService struct {
	ctx       context.Context
	cfg       *Config
	ethClient client.EthClientInterface

	m subsystem.Manager
}

// New ... Initializer
func New(ctx context.Context, cfg *Config, m subsystem.Manager, ethClient client.EthClientInterface) *PessimismService {
	return &PessimismService{
		ctx:       ctx,
		cfg:       cfg,
		ethClient: ethClient,

		m: m,
	}
}
