package engine

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

type Manager struct {
	transit chan core.TransitData
	engine  RiskEngine
	store   *InvariantStore
}

func NewManager() *Manager {
	return &Manager{
		engine:  NewHardCodedEngine(),
		transit: core.NewTransitChannel(),
		store:   NewInvariantStore(),
	}
}

func (em *Manager) Transit() chan core.TransitData {
	return em.transit
}

func (em *Manager) DeployInvariantSession(n core.Network, it core.InvariantType, pt core.PipelineType, invParams any) (core.InvariantUUID, error) {
	inv, err := registry.GetInvariant(it, invParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	invID := core.MakeInvariantUUID(n, pt, it)
	rID := core.MakeRegisterPID(pt, inv.InputType())

	err = em.store.AddInvariant(invID, rID, inv)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	return invID, nil
}

func (em *Manager) EventLoop(ctx context.Context) error {
	logger := logging.WithContext(ctx)

	for {

		select {
		case data := <-em.transit:
			rID := data.GetRegisterPID()
			invs, err := em.store.GetInvariantsByRegisterPID(rID)
			if err != nil {
				logger.Error("Could not find invariants for register ID",
					zap.String("register_id", rID.String()),
					zap.Error(err),
				)
			}

			err = em.engine.Execute(ctx, data, invs)
			if err != nil {
				logger.Error("Could not execute invariants for register ID",
					zap.String("register_id", rID.String()),
					zap.Error(err),
				)
			}

		case <-ctx.Done():
			logging.WithContext(ctx).Debug("Manager received shutdown signal")
			return nil
		}

	}

}
