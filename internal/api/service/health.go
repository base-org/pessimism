package service

import (
	"time"

	"github.com/base-org/pessimism/internal/api/models"
)

// CheckHealth ... Returns health check for server
func (svc *PessimismService) CheckHealth() *models.HealthCheck {
	return &models.HealthCheck{Timestamp: time.Now(), Healthy: true}
}
