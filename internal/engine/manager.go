//go:generate mockgen -package mocks --destination ../mocks/engine_manager.go --mock_names Manager=EngineManager . Manager

package engine

import (
	"context"
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/metrics"
	"github.com/base-org/pessimism/internal/state"

	"go.uber.org/zap"
)

type Config struct {
	WorkerCount int
}

// Manager ... Engine manager interface
type Manager interface {
	GetInputType(ht core.HeuristicType) (core.TopicType, error)
	Transit() chan core.HeuristicInput

	DeleteHeuristicSession(core.UUID) (core.UUID, error)
	DeployHeuristic(cfg *heuristic.DeployConfig) (core.UUID, error)

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

	// Used to receive heuristic input from ETL subsystem
	etlIngress chan core.HeuristicInput
	// Used to send alerts to alerting subsystem
	alertEgress chan core.Alert
	// Used to send execution requests to engine worker subscribers
	workerEgress chan ExecInput

	metrics    metrics.Metricer
	engine     RiskEngine
	addressing *AddressMap
	store      SessionStore
	heuristics registry.HeuristicTable
}

// NewManager ... Initializer
func NewManager(ctx context.Context, cfg *Config, engine RiskEngine, addr *AddressMap,
	store SessionStore, it registry.HeuristicTable, alertEgress chan core.Alert) Manager {
	ctx, cancel := context.WithCancel(ctx)

	em := &engineManager{
		ctx:          ctx,
		cancel:       cancel,
		alertEgress:  alertEgress,
		etlIngress:   make(chan core.HeuristicInput),
		workerEgress: make(chan ExecInput),
		engine:       engine,
		addressing:   addr,
		store:        store,
		heuristics:   it,
		metrics:      metrics.WithContext(ctx),
	}

	// Start engine worker pool for concurrent heuristic execution
	// TODO: Add validation checks for worker count
	for i := 0; i < cfg.WorkerCount; i++ {
		logging.WithContext(ctx).Debug("Starting engine worker routine", zap.Int("worker", i))

		engine.AddWorkerIngress(em.workerEgress)
		go engine.EventLoop(ctx)
	}

	return em
}

// Transit ... Returns inter-subsystem transit channel
func (em *engineManager) Transit() chan core.HeuristicInput {
	return em.etlIngress
}

// DeleteHeuristicSession ... Deletes a heuristic session
func (em *engineManager) DeleteHeuristicSession(_ core.UUID) (core.UUID, error) {
	return core.UUID{}, nil
}

func (em *engineManager) updateSharedState(params *core.SessionParams,
	sk *core.StateKey, PathID core.PathID) error {
	err := sk.SetPathID(PathID)
	// PathID already exists in key but is different than the one we want
	if err != nil && sk.PathID != &PathID {
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
				PathID:  &PathID,
			}

			err = state.InsertUnique(em.ctx, innerKey, argStr)
			if err != nil {
				return err
			}
		}
	}

	logging.WithContext(em.ctx).Debug("Setting to state store",
		zap.String(logging.PathIDKey, PathID.String()),
		zap.String(logging.AddrKey, params.Address().String()))

	return nil
}

func (em *engineManager) DeployHeuristic(cfg *heuristic.DeployConfig) (core.UUID, error) {
	reg, exists := em.heuristics[cfg.HeuristicType]
	if !exists {
		return core.UUID{}, fmt.Errorf("heuristic type %s not found", cfg.HeuristicType)
	}

	if reg.PrepareValidate != nil {
		err := reg.PrepareValidate(cfg.Params)
		if err != nil {
			return core.UUID{}, err
		}
	}

	// Build heuristic instance using constructor functions from data topic definitions
	h, err := reg.Constructor(em.ctx, cfg.Params)
	if err != nil {
		return core.UUID{}, err
	}

	id := core.NewUUID()
	err = em.store.AddSession(id, cfg.PathID, h)
	if err != nil {
		return core.UUID{}, err
	}

	// Shared subsystem state management
	if cfg.Stateful {
		err = em.addressing.Insert(cfg.Params.Address(), cfg.PathID, id)
		if err != nil {
			return core.UUID{}, err
		}

		err = em.updateSharedState(cfg.Params, cfg.StateKey, cfg.PathID)
		if err != nil {
			return core.UUID{}, err
		}
	}

	em.metrics.IncActiveHeuristics(cfg.HeuristicType, cfg.Network, cfg.PathID.PathType())

	return id, nil
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
func (em *engineManager) GetInputType(ht core.HeuristicType) (core.TopicType, error) {
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

func (em *engineManager) executeHeuristics(ctx context.Context, data core.HeuristicInput) {
	if data.Input.Addressed() {
		em.executeAddressHeuristics(ctx, data)
	} else {
		em.executeNonAddressHeuristics(ctx, data)
	}
}

func (em *engineManager) executeAddressHeuristics(ctx context.Context, data core.HeuristicInput) {
	logger := logging.WithContext(ctx)

	ids, err := em.addressing.Get(data.Input.Address, data.PathID)
	if err != nil {
		logger.Error("Could not fetch heuristics by address:pipeline",
			zap.Error(err),
			zap.String(logging.PathIDKey, data.PathID.String()))
		return
	}

	for _, sUUID := range ids {
		h, err := em.store.GetInstanceByUUID(sUUID)
		if err != nil {
			logger.Error("Could not find session by heuristic sUUID",
				zap.Error(err),
				zap.String(logging.PathIDKey, sUUID.String()))
			continue
		}

		em.executeHeuristic(ctx, data, h)
	}
}

func (em *engineManager) executeNonAddressHeuristics(ctx context.Context, data core.HeuristicInput) {
	logger := logging.WithContext(ctx)

	ids, err := em.store.GetUUIDsByPathID(data.PathID)
	if err != nil {
		logger.Error("Could not fetch heuristics for pipeline",
			zap.Error(err),
			zap.String(logging.PathIDKey, data.PathID.String()))
	}

	heuristics, err := em.store.GetInstancesByUUIDs(ids)
	if err != nil {
		logger.Error("Could not fetch heuristics for pipeline",
			zap.Error(err),
			zap.String(logging.PathIDKey, data.PathID.String()))
	}

	for _, h := range heuristics { // Execute all heuristics associated with the pipeline
		em.executeHeuristic(ctx, data, h)
	}
}

// executeHeuristic ... Sends heuristic input to engine worker pool for execution
func (em *engineManager) executeHeuristic(ctx context.Context, data core.HeuristicInput, h heuristic.Heuristic) {
	ei := ExecInput{
		ctx: ctx,
		hi:  data,
		h:   h,
	}

	// Send heuristic input to engine worker pool
	em.workerEgress <- ei
}
