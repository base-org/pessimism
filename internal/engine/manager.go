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
	"github.com/ethereum/go-ethereum/common"

	"go.uber.org/zap"
)

// Manager ... Engine manager interface
type Manager interface {
	core.Subsystem

	Transit() chan core.InvariantInput

	// TODO( ) : Session deletion logic
	DeleteInvariantSession(core.InvSessionUUID) (core.InvSessionUUID, error)
	DeployInvariantSession(cfg *invariant.DeployConfig) (core.InvSessionUUID, error)
}

/*
	NOTE - Manager will need to understand
	when pipeline changes occur that require remapping
	invariant sessions to other pipelines
*/

// engineManager ... Engine management abstraction
type engineManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	etlIngress    chan core.InvariantInput
	alertOutgress chan core.Alert

	engine    RiskEngine
	addresser AddressingMap
	store     SessionStore
}

// NewManager ... Initializer
func NewManager(ctx context.Context,
	alertOutgress chan core.Alert) Manager {
	ctx, cancel := context.WithCancel(ctx)

	em := &engineManager{
		ctx:           ctx,
		cancel:        cancel,
		alertOutgress: alertOutgress,
		etlIngress:    make(chan core.InvariantInput),
		engine:        NewHardCodedEngine(),
		addresser:     NewAddressingMap(),
		store:         NewSessionStore(),
	}

	return em
}

// Transit ... Returns inter-subsystem transit channel
func (em *engineManager) Transit() chan core.InvariantInput {
	return em.etlIngress
}

// DeleteInvariantSession ... Deletes an invariant session
func (em *engineManager) DeleteInvariantSession(_ core.InvSessionUUID) (core.InvSessionUUID, error) {
	return core.NilInvariantUUID(), nil
}

// updateSharedState ... Updates the shared state store
// with contextual information about the invariant session
// to the ETL (e.g. address, events)
func (em *engineManager) updateSharedState(invParams core.InvSessionParams,
	register *core.DataRegister, pUUID core.PipelineUUID) error {
	stateStore, err := state.FromContext(em.ctx)
	if err != nil {
		return err
	}

	key := register.StateKey.WithPUUID(pUUID)
	_, err = stateStore.SetSlice(em.ctx, key, invParams.Address())
	if err != nil {
		return err
	}

	if key.Nested { // Nested addressing
		args := invParams.NestedArgs()

		for _, arg := range args {
			key2 := state.MakeKey(register.DataType, invParams.Address(), false).WithPUUID(pUUID)
			_, err = stateStore.SetSlice(em.ctx, key2, arg)
			if err != nil {
				return err
			}
		}
	}

	logging.NoContext().Debug("Setting to state store",
		zap.String(core.PUUIDKey, pUUID.String()),
		zap.String(core.AddrKey, invParams.Address()))

	return nil
}

// DeployInvariantSession ... Deploys an invariant session to be processed by the engine
func (em *engineManager) DeployInvariantSession(cfg *invariant.DeployConfig) (core.InvSessionUUID, error) {
	inv, err := registry.GetInvariant(cfg.InvType, cfg.InvParams)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	sUUID := core.MakeInvSessionUUID(cfg.Network, cfg.PUUID.PipelineType(), cfg.InvType)
	inv.SetSUUID(sUUID)

	err = em.store.AddInvSession(sUUID, cfg.PUUID, inv)
	if err != nil {
		return core.NilInvariantUUID(), err
	}

	if cfg.Register.Addressing {
		gethAddr := common.HexToAddress(cfg.InvParams.Address())

		err = em.addresser.Insert(cfg.PUUID, sUUID, gethAddr)
		if err != nil {
			return core.NilInvariantUUID(), err
		}

		err = em.updateSharedState(cfg.InvParams, cfg.Register, cfg.PUUID)
		if err != nil {
			return core.NilInvariantUUID(), err
		}
	}

	return sUUID, nil
}

// EventLoop ... Event loop for the engine manager
func (em *engineManager) EventLoop() error {
	logger := logging.WithContext(em.ctx)

	for {
		select {
		case data := <-em.etlIngress: // ETL transit
			logger.Debug("Received invariant input",
				zap.String("input", fmt.Sprintf("%+v", data)))

			em.executeInvariants(em.ctx, data)

		case <-em.ctx.Done(): // Shutdown
			logger.Debug("engineManager received shutdown signal")
			return nil
		}
	}
}

func (em *engineManager) Shutdown() error {
	em.cancel()
	return nil
}

// executeInvariants ... Executes all invariants associated with the input etl pipeline
func (em *engineManager) executeInvariants(ctx context.Context, data core.InvariantInput) {
	if data.Input.Addressed() { // Address based invariant
		em.executeAddressInvariants(ctx, data)
	} else { // Non Address based invariant
		em.executeNonAddressInvariants(ctx, data)
	}
}

// executeAddressInvariants ... Executes all address specific invariants associated with the input etl pipeline
func (em *engineManager) executeAddressInvariants(ctx context.Context, data core.InvariantInput) {
	logger := logging.WithContext(ctx)

	sUUID, err := em.addresser.GetSessionUUIDByPair(data.Input.Address, data.PUUID)
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

	// Fetch all invariants associated with the pipeline
	sUUIDs, err := em.store.GetInvSessionsForPipeline(data.PUUID)
	if err != nil {
		logger.Error("Could not fetch invariants for pipeline",
			zap.Error(err),
			zap.String(core.PUUIDKey, data.PUUID.String()))
	}

	// Fetch all invariants by SUUIDs
	invs, err := em.store.GetInvariantsByUUIDs(sUUIDs...)
	if err != nil {
		logger.Error("Could not fetch invariants for pipeline",
			zap.Error(err),
			zap.String(core.PUUIDKey, data.PUUID.String()))
	}

	for _, inv := range invs { // Execute all invariants associated with the pipeline
		em.executeInvariant(ctx, data, inv)
	}
}

// executeInvariant ... Executes a single invariant using the risk engine
func (em *engineManager) executeInvariant(ctx context.Context, data core.InvariantInput, inv invariant.Invariant) {
	logger := logging.WithContext(ctx)

	// Execute invariant using risk engine and return alert if invalidation occurs
	outcome, invalid := em.engine.Execute(ctx, data.Input, inv)

	if invalid {
		alert := core.Alert{
			Timestamp: outcome.TimeStamp,
			SUUID:     inv.SUUID(),
			Content:   outcome.Message,
		}

		logger.Warn("Invariant alert",
			zap.String(core.SUUIDKey, inv.SUUID().String()),
			zap.String("message", outcome.Message))

		em.alertOutgress <- alert
	}
}
