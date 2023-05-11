package service

import (
	"context"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/etl/pipeline"
)

type Config struct {
	L1RpcEndpoint string
	L2RpcEndpoint string
}

func (cfg *Config) GetEndpointFromNetwork(n core.Network) (string, error) {
	switch n {
	case core.Layer1:
		return cfg.L1RpcEndpoint, nil

	case core.Layer2:
		return cfg.L2RpcEndpoint, nil

	default:
		return "", fmt.Errorf("could not find endpoint for network %s", n.String())
	}
}

// Service ...
type Service interface {
	ProcessInvariantRequest(ir models.InvRequestBody) (core.InvSessionUUID, error)
	CheckHealth() *models.HealthCheck
}

// PessimismService ...
type PessimismService struct {
	ctx           context.Context
	cfg           *Config
	etlManager    *pipeline.Manager
	engineManager *engine.Manager
}

// New ... Initializer
func New(ctx context.Context, cfg *Config, etlManager *pipeline.Manager,
	engineManager *engine.Manager) *PessimismService {
	return &PessimismService{
		ctx:           ctx,
		cfg:           cfg,
		etlManager:    etlManager,
		engineManager: engineManager,
	}
}

// CheckHealth ... Returns health check for server
func (svc *PessimismService) CheckHealth() *models.HealthCheck {
	return &models.HealthCheck{Timestamp: time.Now(), Healthy: true}
}
