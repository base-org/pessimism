package service

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

// CheckHealth ... Returns health check for server
func (svc *PessimismService) CheckHealth() *models.HealthCheck {
	var wg sync.WaitGroup

	checks := 2 // number of checks to perform
	wg.Add(checks)

	hc := &models.ChainConnectionStatus{
		IsL1Healthy: false,
		IsL2Healthy: false,
	}

	go func() {
		defer wg.Done()

		hc.IsL1Healthy = svc.CheckETHRPCHealth(core.Layer1)
	}()

	go func() {
		defer wg.Done()
		hc.IsL2Healthy = svc.CheckETHRPCHealth(core.Layer2)
	}()

	wg.Wait()

	return &models.HealthCheck{
		Timestamp:             time.Now(),
		Healthy:               hc.IsL1Healthy && hc.IsL2Healthy,
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
