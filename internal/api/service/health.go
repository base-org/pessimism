package service

import (
	"time"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// CheckHealth ... Returns health check for server
func (svc *PessimismService) CheckHealth() *models.HealthCheck {
	healthCheck := &models.NodeConnectionStatus{}

	healthCheck.IsL1Healthy = svc.CheckETHRPCHealth(svc.cfg.L1RpcEndpoint)
	healthCheck.IsL2Healthy = svc.CheckETHRPCHealth(svc.cfg.L2RpcEndpoint)

	healthy := healthCheck.IsL1Healthy && healthCheck.IsL2Healthy

	return &models.HealthCheck{Timestamp: time.Now(), Healthy: healthy}
}

func (svc *PessimismService) CheckETHRPCHealth(url string) bool {
	logger := logging.WithContext(svc.ctx)

	err := svc.ethClient.DialContext(svc.ctx, url)
	if err != nil {
		logger.Error("error conntecting to %s", zap.String("url", url))
		return false
	}

	_, err = svc.ethClient.HeaderByNumber(svc.ctx, nil)
	if err != nil {
		logger.Error("error connecting to url", zap.String("url", url))
		return false
	}

	logger.Debug("successfully connected", zap.String("url", url))
	return true
}
