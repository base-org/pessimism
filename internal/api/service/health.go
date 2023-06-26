package service

import (
	"time"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"

	"go.uber.org/zap"
)

// CheckHealth ... Returns health check for server
func (svc *PessimismService) CheckHealth() *models.HealthCheck {
	// TODO(#88): Parallelized Node Queries for Health Checking
	hc := &models.ChainConnectionStatus{
		IsL1Healthy: svc.CheckETHRPCHealth(core.Layer1),
		IsL2Healthy: svc.CheckETHRPCHealth(core.Layer2),
	}

	healthy := hc.IsL1Healthy && hc.IsL2Healthy

	return &models.HealthCheck{
		Timestamp:             time.Now(),
		Healthy:               healthy,
		ChainConnectionStatus: hc,
	}
}

func (svc *PessimismService) CheckETHRPCHealth(n core.Network) bool {
	logger := logging.WithContext(svc.ctx)
	ethClient, err := client.FromContext(svc.ctx, n)
	if err != nil {
		logger.Error("error getting client from context", zap.Error(err))
		return false
	}

	_, err = ethClient.HeaderByNumber(svc.ctx, nil)
	if err != nil {
		logger.Error("error connecting to client", zap.String("network", n.String()))
		return false
	}

	logger.Debug("successfully connected", zap.String("network", n.String()))
	return true
}
