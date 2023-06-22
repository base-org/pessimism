package service

import (
	"time"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// CheckHealth ... Returns health check for server
func (svc *PessimismService) CheckHealth() *models.HealthCheck {
	hc := &models.ChainConnectionStatus{
		IsL1Healthy: svc.CheckETHRPCHealth(svc.cfg.L1RpcEndpoint),
		IsL2Healthy: svc.CheckETHRPCHealth(svc.cfg.L2RpcEndpoint),
	}

	healthy := hc.IsL1Healthy && hc.IsL2Healthy

	return &models.HealthCheck{
		Timestamp:             time.Now(),
		Healthy:               healthy,
		ChainConnectionStatus: *hc,
	}
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
