package engine

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

/*
	NOTE - Manager will need to understand
	when pipeline changes occur that require remapping
	invariant sessions to other pipelines
*/

// Manager ... Engine management abstraction
type Manager struct {
	transit chan core.InvariantInput

	engine RiskEngine
	store  SessionStore
}

// NewManager ... Initializer
func NewManager() (*Manager, func()) {
	m := &Manager{
		engine:  NewHardCodedEngine(),
		transit: make(chan core.InvariantInput),
		store:   NewSessionStore(),
	}

	shutDown := func() {
		close(m.transit)
	}

	return m, shutDown
}

// Transit ... Returns inter-subsystem transit channel
func (em *Manager) Transit() chan core.InvariantInput {
	return em.transit
}

// TODO() :
// DeleteInvariantSession ...
func (em *Manager) DeleteInvariantSession(_ core.InvSessionUUID) (core.InvSessionUUID, error) {
	return core.NilInvariantUUID(), nil
}

// DeployInvariantSession ... Deploys an
func (em *Manager) DeployInvariantSession(n core.Network, pUUUID core.PipelineUUID, it core.InvariantType,
	pt core.PipelineType, invParams any) (core.InvSessionUUID, error) {
	inv, err := registry.GetInvariant(it, invParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	sessionID := core.MakeInvSessionUUID(n, pt, it)
	inv.WithUUID(sessionID)

	err = em.store.AddInvSession(sessionID, pUUUID, inv)
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
			logger.Debug("Received invariant input",
				zap.String("input", fmt.Sprintf("%+v", data)))

			em.executeInvariants(ctx, data)

		case <-ctx.Done():
			logger.Debug("engineManager received shutdown signal")
			return nil
		}
	}
}

// executeInvariants ... Executes all invariants associated with the input etl pipeline
func (em *Manager) executeInvariants(ctx context.Context, data core.InvariantInput) {
	logger := logging.WithContext(ctx)

	invUUIDs, err := em.store.GetInvSessionsForPipeline(data.PUUID) // TODO - Change to Get method
	if err != nil {
		logger.Error("Could not fetch invariants for pipeline",
			zap.Error(err),
			zap.String("puuid", data.PUUID.String()))
	}

	invs, err := em.store.GetInvariantsByUUIDs(invUUIDs...)
	if err != nil {
		logger.Error("Could not fetch invariants for pipeline",
			zap.Error(err),
			zap.String("puuid", data.PUUID.String()))
	}

	for i, inv := range invs {
		sUUID := invUUIDs[i]

		err = em.engine.Execute(ctx, data.Input, inv)
		if err != nil {
			logger.Error("Could not execute invariant",
				zap.String("puuid", data.PUUID.String()),
				zap.String("session_uuid", sUUID.String()),
			)
		}
	}
}
