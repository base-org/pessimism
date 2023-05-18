//go:generate mockgen -package mocks --destination ../mocks/engine_manager.go --mock_names Manager=EngineManager . Manager

package engine

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"

	"go.uber.org/zap"
)

type Manager interface {
	Transit() chan core.InvariantInput
	EventLoop(ctx context.Context) error

	// TODO( ) : Session deletion logic
	DeleteInvariantSession(_ core.InvSessionUUID) (core.InvSessionUUID, error)
	DeployInvariantSession(n core.Network, pUUUID core.PipelineUUID, it core.InvariantType,
		pt core.PipelineType, invParams core.InvSessionParams) (core.InvSessionUUID, error)
}

/*
	NOTE - Manager will need to understand
	when pipeline changes occur that require remapping
	invariant sessions to other pipelines
*/

// Manager ... Engine management abstraction
type engineManager struct {
	ctx          context.Context
	etlTransit   chan core.InvariantInput
	alertTransit chan core.Alert

	engine    RiskEngine
	addresser AddressingMap
	store     SessionStore
}

// NewManager ... Initializer
func NewManager(ctx context.Context,
	alertTransit chan core.Alert) (Manager, func()) {
	em := &engineManager{
		ctx:        ctx,
		etlTransit: make(chan core.InvariantInput),
		engine:     NewHardCodedEngine(),
		addresser:  NewAddressingMap(),
		store:      NewSessionStore(),
	}

	shutDown := func() {
		close(em.etlTransit)
	}

	return em, shutDown
}

// Transit ... Returns inter-subsystem transit channel
func (em *engineManager) Transit() chan core.InvariantInput {
	return em.etlTransit
}

// TODO() :
// DeleteInvariantSession ...
func (em *engineManager) DeleteInvariantSession(_ core.InvSessionUUID) (core.InvSessionUUID, error) {
	return core.NilInvariantUUID(), nil
}

// DeployInvariantSession ...
func (em *engineManager) DeployInvariantSession(n core.Network, pUUUID core.PipelineUUID, it core.InvariantType,
	pt core.PipelineType, invParams core.InvSessionParams) (core.InvSessionUUID, error) {
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

	if inv.Addressing() {
		stateStore, err := state.FromContext(em.ctx)
		if err != nil {
			return core.NilInvariantUUID(), err
		}

		logging.NoContext().Debug("Setting to state store",
			zap.String(core.PUUIDKey, pUUUID.String()),
			zap.String(core.AddrKey, invParams.Address()))

		stateStore.Set(em.ctx, pUUUID.String(), invParams.Address())
	}

	return sessionID, nil
}

// EventLoop ...
func (em *engineManager) EventLoop(ctx context.Context) error {
	logger := logging.WithContext(ctx)

	for {
		select {
		case data := <-em.etlTransit:
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
func (em *engineManager) executeInvariants(ctx context.Context, data core.InvariantInput) {

	if data.Input.Address != nil { // Address based invariant
		em.executeAddressInvariants(ctx, data)

	} else { // Non Address based invariant
		em.executeNonAddressInvariants(ctx, data)
	}

}

func (em *engineManager) executeAddressInvariants(ctx context.Context, data core.InvariantInput) {
	logger := logging.WithContext(ctx)

	sUUID, err := em.addresser.GetSessionUUIDByPair(*data.Input.Address, data.PUUID)
	if err != nil {
		logger.Error("Could not fetch invariants by address:pipeline",
			zap.Error(err),
			zap.String(core.PUUIDKey, data.PUUID.String()))
		return
	}

	inv, err := em.store.GetInvSessionByUUID(sUUID)
	if err != nil {
		logger.Error("Could not session by invariant sUUID",
			zap.Error(err),
			zap.String(core.PUUIDKey, sUUID.String()))
		return
	}

	em.executeInvariant(ctx, data, inv)

}

func (em *engineManager) executeInvariant(ctx context.Context, data core.InvariantInput, inv invariant.Invariant) {
	logger := logging.WithContext(ctx)

	alert, err := em.engine.Execute(ctx, data.Input, inv)
	if err != nil {
		logger.Error("Could not execute invariant",
			zap.String(core.PUUIDKey, data.PUUID.String()),
			zap.String(core.SUUIDKey, inv.UUID().String()),
			zap.String(core.AddrKey, data.Input.Address.String()))
		return
	}

	if alert != nil {
		logger.Error("Invariant alert")
		em.alertTransit <- *alert
	}

}

func (em *engineManager) executeNonAddressInvariants(ctx context.Context, data core.InvariantInput) {
	logger := logging.WithContext(ctx)

	invUUIDs, err := em.store.GetInvSessionsForPipeline(data.PUUID)
	if err != nil {
		logger.Error("Could not fetch invariants for pipeline",
			zap.Error(err),
			zap.String(core.PUUIDKey, data.PUUID.String()))
	}

	invs, err := em.store.GetInvariantsByUUIDs(invUUIDs...)
	if err != nil {
		logger.Error("Could not fetch invariants for pipeline",
			zap.Error(err),
			zap.String(core.PUUIDKey, data.PUUID.String()))
	}

	for _, inv := range invs {
		em.executeInvariant(ctx, data, inv)
	}
}
