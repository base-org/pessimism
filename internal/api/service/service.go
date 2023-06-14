//go:generate mockgen -package mocks --destination ../../mocks/api_service.go . Service

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/etl/pipeline"
)

// Config ... Used to store necessary API service config values
type Config struct {
	L1RpcEndpoint  string
	L2RpcEndpoint  string
	L1PollInterval int
	L2PollInterval int
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
	ProcessInvariantRequest(ir models.InvRequestBody) (core.SUUID, error)
	RunInvariantSession(params models.InvRequestParams) (core.SUUID, error)

	CheckHealth() *models.HealthCheck
}

// PessimismService ... API service
type PessimismService struct {
	ctx context.Context
	cfg *Config

	alertManager  alert.Manager
	etlManager    pipeline.Manager
	engineManager engine.Manager
}

// New ... Initializer
func New(ctx context.Context, cfg *Config, alertManager alert.Manager, etlManager pipeline.Manager,
	engineManager engine.Manager) *PessimismService {
	return &PessimismService{
		ctx: ctx,
		cfg: cfg,

		alertManager:  alertManager,
		etlManager:    etlManager,
		engineManager: engineManager,
	}
}
