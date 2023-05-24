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

// Manager ... Engine manager interface
type Manager interface {
	Transit() chan core.InvariantInput
	EventLoop(ctx context.Context) error

	// TODO( ) : Session deletion logic
	DeleteInvariantSession(core.InvSessionUUID) (core.InvSessionUUID, error)
	DeployInvariantSession(core.Network, core.PipelineUUID, core.InvariantType,
		core.PipelineType, core.InvSessionParams) (core.InvSessionUUID, error)
}

/*
	NOTE - Manager will need to understand
	when pipeline changes occur that require remapping
	invariant sessions to other pipelines
*/

// engineManager ... Engine management abstraction
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
		ctx:          ctx,
		alertTransit: alertTransit,
		etlTransit:   make(chan core.InvariantInput),
		engine:       NewHardCodedEngine(),
		addresser:    NewAddressingMap(),
		store:        NewSessionStore(),
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
// DeleteInvariantSession ... Deletes an invariant session
func (em *engineManager) DeleteInvariantSession(_ core.InvSessionUUID) (core.InvSessionUUID, error) {
	return core.NilInvariantUUID(), nil
}

// DeployInvariantSession ... Deploys an invariant session to be processed by the engine
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

	if inv.Addressing() { // Address based invariant
		stateStore, err := state.FromContext(em.ctx)
		if err != nil {
			return core.NilInvariantUUID(), err
		}

		logging.NoContext().Debug("Setting to state store",
			zap.String(core.PUUIDKey, pUUUID.String()),
			zap.String(core.AddrKey, invParams.Address()))

		// Set address to shared state store for the pipeline to utilize
		_, err = stateStore.Set(em.ctx, pUUUID.String(), invParams.Address())
		if err != nil {
			return core.NilInvariantUUID(), err
		}
	}

	return sessionID, nil
}

// EventLoop ... Event loop for the engine manager
func (em *engineManager) EventLoop(ctx context.Context) error {
	logger := logging.WithContext(ctx)

	for {
		select {
		case data := <-em.etlTransit: // ETL transit
			logger.Debug("Received invariant input",
				zap.String("input", fmt.Sprintf("%+v", data)))

			em.executeInvariants(ctx, data)

		case <-ctx.Done(): // Shutdown
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

// executeAddressInvariants ... Executes all address specific invariants associated with the input etl pipeline
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

// executeNonAddressInvariants ... Executes all non address specific invariants associated with the input etl pipeline
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

// executeInvariant ... Executes a single invariant using the risk engine
func (em *engineManager) executeInvariant(ctx context.Context, data core.InvariantInput, inv invariant.Invariant) {
	logger := logging.WithContext(ctx)

	// Execute invariant using risk engine and return alert if invalidation occurs
	alert, invalid := em.engine.Execute(ctx, data.Input, inv)

	if invalid {
		logger.Warn("Invariant alert", zap.String(core.SUUIDKey, inv.UUID().String()))
		em.alertTransit <- *alert
	}
}
