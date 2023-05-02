package service

import (
	"time"

	"github.com/base-org/pessimism/internal/api/models"
)

// Service ...
type Service interface {
	CheckHealth() *models.HealthCheck
}

// PessimismService ...
type PessimismService struct {
}

// New ... Initializer
func New() *PessimismService {
	return &PessimismService{}
}

// CheckHealth ... Returns health check for server
func (svc *PessimismService) CheckHealth() *models.HealthCheck {
	return &models.HealthCheck{Timestamp: time.Now(), Healthy: true}
}
