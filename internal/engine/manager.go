package engine

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

type Manager struct {
	closeChan chan int
	transit   chan core.InvariantInput
	engine    RiskEngine
	store     *InvariantStore
}

func NewManager() (*Manager, func()) {
	m := &Manager{
		engine:    NewHardCodedEngine(),
		closeChan: make(chan int, 1),
		transit:   make(chan core.InvariantInput),
		store:     NewInvariantStore(),
	}

	shutDown := func() {
		close(m.transit)
		m.closeChan <- 0
	}

	return m, shutDown
}

func (em *Manager) Transit() chan core.InvariantInput {
	return em.transit
}

func (em *Manager) DeployInvariantSession(n core.Network, pUUUID core.PipelineUUID, it core.InvariantType,
	pt core.PipelineType, invParams any) (core.InvSessionUUID, error) {
	inv, err := registry.GetInvariant(it, invParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	sessionID := core.MakeInvSessionUUID(n, pt, it)

	err = em.store.AddInvariant(sessionID, pUUUID, inv)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	return sessionID, nil
}

func (em *Manager) EventLoop(ctx context.Context) error {
	logger := logging.WithContext(ctx)

	for {
		select {
		case data := <-em.transit:

			invIDs := em.store.invPipeLineMap[data.PUUID] // TODO - Change to Get method
			if len(invIDs) == 0 {
				logging.WithContext(ctx).
					Error("No invariants found", zap.String("puuid", data.PUUID.String()))
			}

			invs := make([]invariant.Invariant, len(invIDs))

			for i, id := range invIDs {
				invs[i] = em.store.invMap[id]
			}

			err := em.engine.Execute(ctx, data.Input, invs)
			if err != nil {
				logger.Error("Could not execute invariants for register ID",
					zap.Error(err),
				)
			}

		case <-em.closeChan:
			logging.WithContext(ctx).Debug("Engine manager received shutdown signal")
			return nil
		}
	}
}
