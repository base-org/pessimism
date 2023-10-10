//go:generate mockgen -package mocks --destination ../../mocks/api_service.go . Service

package service

import (
	"context"

	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/subsystem"
)

// Service ... Interface for API service
type Service interface {
	ProcessHeuristicRequest(ir *models.SessionRequestBody) (core.SUUID, error)
	RunHeuristicSession(params *models.SessionRequestParams) (core.SUUID, error)

	CheckHealth() *models.HealthCheck
	CheckETHRPCHealth(n core.Network) bool
}

// PessimismService ... API service
type PessimismService struct {
	ctx context.Context
	m   subsystem.Subsystem
}

// New ... Initializer
func New(ctx context.Context, m subsystem.Subsystem) *PessimismService {
	return &PessimismService{
		ctx: ctx,
		m:   m,
	}
}
