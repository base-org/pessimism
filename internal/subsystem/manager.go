package subsystem

import (
	"context"
	"sync"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/base-org/pessimism/internal/etl/pipeline"
	"github.com/base-org/pessimism/internal/logging"
	"go.uber.org/zap"
)

// Manager ... Subsystem manager struct
type Manager struct {
	etl  pipeline.Manager
	eng  engine.Manager
	alrt alert.AlertingManager

	*sync.WaitGroup
}

// NewManager ... Initializer for the subsystem manager
func NewManager(etl pipeline.Manager, eng engine.Manager,
	alrt alert.AlertingManager,
) *Manager {
	return &Manager{
		etl:       etl,
		eng:       eng,
		alrt:      alrt,
		WaitGroup: &sync.WaitGroup{},
	}
}

func (m *Manager) Shutdown() error {
	if err := m.etl.Shutdown(); err != nil {
		return err
	}

	if err := m.eng.Shutdown(); err != nil {
		return err
	}

	if err := m.alrt.Shutdown(); err != nil {
		return err
	}

	return nil
}

// StartEventRoutines ... Starts the event loop routines for the subsystems
func (m *Manager) StartEventRoutines(ctx context.Context) {
	logger := logging.WithContext(ctx)

	m.Add(1)
	go func() { // EngineManager driver thread
		defer m.Done()

		if err := m.eng.EventLoop(); err != nil {
			logger.Error("engine manager event loop error", zap.Error(err))
		}
	}()

	m.Add(1)
	go func() { // AlertManager driver thread
		defer m.Done()

		if err := m.alrt.EventLoop(); err != nil {
			logger.Error("alert manager event loop error", zap.Error(err))
		}
	}()

	m.Add(1)
	go func() { // ETL driver thread
		defer m.Done()

		if err := m.alrt.EventLoop(); err != nil {
			logger.Error("ETL manager event loop error", zap.Error(err))
		}
	}()
}
