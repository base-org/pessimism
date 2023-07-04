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
	ProcessInvariantRequest(ir *models.InvRequestBody) (core.SUUID, error)
	RunInvariantSession(params *models.InvRequestParams) (core.SUUID, error)

	CheckHealth() *models.HealthCheck
	CheckETHRPCHealth(n core.Network) bool
}

// PessimismService ... API service
type PessimismService struct {
	ctx context.Context
	m   subsystem.Manager
}

// New ... Initializer
func New(ctx context.Context, m subsystem.Manager) *PessimismService {
	return &PessimismService{
		ctx: ctx,
		m:   m,
	}
}
