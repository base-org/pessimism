//go:generate mockgen -package mocks --destination ../mocks/engine_manager.go --mock_names Manager=EngineManager . Manager

package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/state"

	"go.uber.org/zap"
)

// Manager ... Engine manager interface
type Manager interface {
	GetInputType(ht core.HeuristicType) (core.RegisterType, error)
	Transit() chan core.HeuristicInput

	DeleteHeuristicSession(core.SUUID) (core.SUUID, error)
	DeployHeuristicSession(cfg *heuristic.DeployConfig) (core.SUUID, error)

	core.Subsystem
}

/*
	NOTE - Manager will need to understand
	when pipeline changes occur that require remapping
	heuristic sessions to other pipelines
*/

// engineManager ... Engine management abstraction
type engineManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	etlIngress    chan core.HeuristicInput
	alertOutgress chan core.Alert

	metrics    metrics.Metricer
	engine     RiskEngine
	addresser  AddressingMap
	store      SessionStore
	heuristics registry.HeuristicTable
}

// NewManager ... Initializer
func NewManager(ctx context.Context, engine RiskEngine, addr AddressingMap,
	store SessionStore, it registry.HeuristicTable, alertOutgress chan core.Alert) Manager {
	ctx, cancel := context.WithCancel(ctx)

	em := &engineManager{
		ctx:           ctx,
		cancel:        cancel,
		alertOutgress: alertOutgress,
		etlIngress:    make(chan core.HeuristicInput),
		engine:        engine,
		addresser:     addr,
		store:         store,
		heuristics:    it,
		metrics:       metrics.WithContext(ctx),
	}

	return em
}

// Transit ... Returns inter-subsystem transit channel
func (em *engineManager) Transit() chan core.HeuristicInput {
	return em.etlIngress
}

// DeleteHeuristicSession ... Deletes a heuristic session
func (em *engineManager) DeleteHeuristicSession(_ core.SUUID) (core.SUUID, error) {
	return core.NilSUUID(), nil
}

// updateSharedState ... Updates the shared state store
// with contextual information about the heuristic session
// to the ETL (e.g. address, events)
func (em *engineManager) updateSharedState(params *core.SessionParams,
	sk *core.StateKey, pUUID core.PUUID) error {
	err := sk.SetPUUID(pUUID)
	// PUUID already exists in key but is different than the one we want
	if err != nil && sk.PUUID != &pUUID {
		return err
	}

	// Use accessor method to insert entry into state store
	err = state.InsertUnique(em.ctx, sk, params.Address().String())
	if err != nil {
		return err
	}

	if sk.IsNested() { // Nested addressing
		for _, arg := range params.NestedArgs() {
			argStr, success := arg.(string)
			if !success {
				return fmt.Errorf("invalid event string")
			}

			// Build nested key
			innerKey := &core.StateKey{
				Nesting: false,
				Prefix:  sk.Prefix,
				ID:      params.Address().String(),
				PUUID:   &pUUID,
			}

			err = state.InsertUnique(em.ctx, innerKey, argStr)
			if err != nil {
				return err
			}
		}
	}

	logging.WithContext(em.ctx).Debug("Setting to state store",
		zap.String(logging.PUUIDKey, pUUID.String()),
		zap.String(logging.AddrKey, params.Address().String()))

	return nil
}

// DeployHeuristicSession ... Deploys a heuristic session to be processed by the engine
func (em *engineManager) DeployHeuristicSession(cfg *heuristic.DeployConfig) (core.SUUID, error) {
	reg, exists := em.heuristics[cfg.HeuristicType]
	if !exists {
		return core.NilSUUID(), fmt.Errorf("heuristic type %s not found", cfg.HeuristicType)
	}

	if reg.PrepareValidate != nil { // Prepare & validate the heuristic params for stateful consumption
		err := reg.PrepareValidate(cfg.Params)
		if err != nil {
			return core.NilSUUID(), err
		}
	}

	// Build heuristic instance using constructor function from register definition
	h, err := reg.Constructor(em.ctx, cfg.Params)
	if err != nil {
		return core.NilSUUID(), err
	}

	// Generate session UUID and set it to the heuristic
	sUUID := core.MakeSUUID(cfg.Network, cfg.PUUID.PipelineType(), cfg.HeuristicType)
	h.SetSUUID(sUUID)

	err = em.store.AddSession(sUUID, cfg.PUUID, h)
	if err != nil {
		return core.NilSUUID(), err
	}

	// Shared subsystem state management
	if cfg.Stateful {
		err = em.addresser.Insert(cfg.Params.Address(), cfg.PUUID, sUUID)
		if err != nil {
			return core.NilSUUID(), err
		}

		err = em.updateSharedState(cfg.Params, cfg.StateKey, cfg.PUUID)
		if err != nil {
			return core.NilSUUID(), err
		}
	}

	em.metrics.IncActiveHeuristics(cfg.HeuristicType, cfg.Network, cfg.PUUID.PipelineType())

	return sUUID, nil
}

// EventLoop ... Event loop for the engine manager
func (em *engineManager) EventLoop() error {
	logger := logging.WithContext(em.ctx)

	for {
		select {
		case data := <-em.etlIngress: // ETL transit
			logger.Debug("Received heuristic input",
				zap.String("input", fmt.Sprintf("%+v", data)))

			em.executeHeuristics(em.ctx, data)

		case <-em.ctx.Done(): // Shutdown
			logger.Debug("engineManager received shutdown signal")
			return nil
		}
	}
}

// GetInputType ... Returns the register input type for the heuristic type
func (em *engineManager) GetInputType(ht core.HeuristicType) (core.RegisterType, error) {
	val, exists := em.heuristics[ht]
	if !exists {
		return 0, fmt.Errorf("heuristic type %s not found", ht)
	}

	return val.InputType, nil
}

// Shutdown ... Shuts down the engine manager
func (em *engineManager) Shutdown() error {
	em.cancel()
	return nil
}

// executeHeuristics ... Executes all heuristics associated with the input etl pipeline
func (em *engineManager) executeHeuristics(ctx context.Context, data core.HeuristicInput) {
	if data.Input.Addressed() { // Address based heuristic
		em.executeAddressHeuristics(ctx, data)
	} else { // Non Address based heuristic
		em.executeNonAddressHeuristics(ctx, data)
	}
}

// executeAddressHeuristics ... Executes all address specific heuristics associated with the input etl pipeline
func (em *engineManager) executeAddressHeuristics(ctx context.Context, data core.HeuristicInput) {
	logger := logging.WithContext(ctx)

	ids, err := em.addresser.GetSUUIDsByPair(data.Input.Address, data.PUUID)
	if err != nil {
		logger.Error("Could not fetch heuristics by address:pipeline",
			zap.Error(err),
			zap.String(logging.PUUIDKey, data.PUUID.String()))
		return
	}

	for _, sUUID := range ids {
		h, err := em.store.GetInstanceByUUID(sUUID)
		if err != nil {
			logger.Error("Could not session by heuristic sUUID",
				zap.Error(err),
				zap.String(logging.PUUIDKey, sUUID.String()))
			continue
		}

		em.executeHeuristic(ctx, data, h)
	}
}

// executeNonAddressHeuristics ... Executes all non address specific heuristics associated with the input etl pipeline
func (em *engineManager) executeNonAddressHeuristics(ctx context.Context, data core.HeuristicInput) {
	logger := logging.WithContext(ctx)

	// Fetch all session UUIDs associated with the pipeline
	sUUIDs, err := em.store.GetSUUIDsByPUUID(data.PUUID)
	if err != nil {
		logger.Error("Could not fetch heuristics for pipeline",
			zap.Error(err),
			zap.String(logging.PUUIDKey, data.PUUID.String()))
	}

	// Fetch all heuristics for a slice of SUUIDs
	heuristics, err := em.store.GetInstancesByUUIDs(sUUIDs)
	if err != nil {
		logger.Error("Could not fetch heuristics for pipeline",
			zap.Error(err),
			zap.String(logging.PUUIDKey, data.PUUID.String()))
	}

	for _, h := range heuristics { // Execute all heuristics associated with the pipeline
		em.executeHeuristic(ctx, data, h)
	}
}

// executeHeuristic ... Executes a single heuristic using the risk engine
func (em *engineManager) executeHeuristic(ctx context.Context, data core.HeuristicInput, h heuristic.Heuristic) {
	logger := logging.WithContext(ctx)

	start := time.Now()
	// 1.  Executes heuristic for outcomes using risk engine
	outcomes, activated := em.engine.Execute(ctx, data.Input, h)

	em.metrics.RecordHeuristicRun(h)
	em.metrics.RecordInvExecutionTime(h, float64(time.Since(start).Nanoseconds()))

	// 2. Send alerts if activated
	if activated {

		for _, oc := range outcomes {
			// Generate & send alert
			alert := core.Alert{
				Timestamp: oc.TimeStamp,
				SUUID:     h.SUUID(),
				Content:   oc.Message,
				PUUID:     data.PUUID,
				Ptype:     data.PUUID.PipelineType(),
			}

			logger.Warn("Heuristic alert",
				zap.String(logging.SUUIDKey, h.SUUID().String()),
				zap.String("message", oc.Message))

			em.alertOutgress <- alert
		}
	}
}
