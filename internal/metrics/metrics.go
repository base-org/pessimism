package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"go.uber.org/zap"
)

const (
	metricsNamespace    = "pessimism"
	SubsystemInvariants = "invariants"
	SubsystemEtl        = "etl"
)

const serverShutdownTimeout = 10 * time.Second

type Config struct {
	Host              string
	Port              int
	Enabled           bool
	ReadHeaderTimeout int
}

type Metricer interface {
	IncActiveInvariants()
	DecActiveInvariants()
	IncActivePipelines()
	DecActivePipelines()
	RecordInvariantRun(invariant string)
	RecordAlertGenerated(invariant string)
	RecordNodeError(node string)
	RecordUp()
	Start()
	Shutdown(ctx context.Context) error
	Document() []DocumentedMetric
}

type Metrics struct {
	ActiveInvariants prometheus.Gauge
	ActivePipelines  prometheus.Gauge
	Up               prometheus.Gauge
	InvariantRuns    *prometheus.CounterVec
	AlarmsGenerated  *prometheus.CounterVec
	NodeErrors       *prometheus.CounterVec

	registry *prometheus.Registry
	factory  Factory
	server   *http.Server
}

var stats Metricer = new(noopMetricer)

// WithContext returns a Metricer from the given context. If no Metricer is found,
// the default noopMetricer is returned.
func WithContext(ctx context.Context) Metricer {
	if ctx == nil {
		return stats
	}

	if ctxStats, ok := ctx.Value(core.Metrics).(Metricer); ok {
		return ctxStats
	}

	return stats
}

// New ... Creates a new metrics server registered with defined custom metrics
func New(ctx context.Context, cfg *Config) (Metricer, func(), error) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	factory := With(registry)

	stats = &Metrics{
		Up: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "up",
			Help:      "1 if the service is up",
		}),
		ActiveInvariants: factory.NewGauge(prometheus.GaugeOpts{
			Name:      "active_invariants",
			Help:      "Number of active invariants",
			Namespace: metricsNamespace,
			Subsystem: SubsystemInvariants,
		}),

		ActivePipelines: factory.NewGauge(prometheus.GaugeOpts{
			Name:      "active_pipelines",
			Help:      "Number of active pipelines",
			Namespace: metricsNamespace,
			Subsystem: SubsystemEtl,
		}),

		InvariantRuns: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "invariant_runs_total",
			Help:      "Number of times a specific invariant has been run",
			Namespace: metricsNamespace,
			Subsystem: SubsystemInvariants,
		}, []string{"invariant"}),

		AlarmsGenerated: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "alarms_generated_total",
			Help:      "Number of total alarms generated for a given invariant",
			Namespace: metricsNamespace,
		}, []string{"invariant"}),

		NodeErrors: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "node_errors_total",
			Help:      "Number of node errors caught",
			Namespace: metricsNamespace,
		}, []string{"node"}),

		registry: registry,
		factory:  factory,
		server:   initServer(cfg, registry),
	}

	stop := func() {
		logging.WithContext(ctx).Info("starting to shutdown metrics server")
		ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		if err := stats.Shutdown(ctx); err != nil {
			logging.WithContext(ctx).Error("failed to shutdown metrics server: %v", zap.Error(err))
			panic(err)
		}
		defer cancel()
	}

	return stats, stop, nil
}

// RecordUp ... Records that the service has been successfully started
func (m *Metrics) RecordUp() {
	prometheus.MustRegister()
	m.Up.Set(1)
}

// IncActiveInvariants ... Increments the number of active invariants
func (m *Metrics) IncActiveInvariants() {
	m.ActiveInvariants.Inc()
}

// DecActiveInvariants ... Decrements the number of active invariants
func (m *Metrics) DecActiveInvariants() {
	m.ActiveInvariants.Dec()
}

// IncActivePipelines ... Increments the number of active pipelines
func (m *Metrics) IncActivePipelines() {
	m.ActivePipelines.Inc()
}

// DecActivePipelines ... Decrements the number of active pipelines
func (m *Metrics) DecActivePipelines() {
	m.ActivePipelines.Dec()
}

// RecordInvariantRun ... Records that a given invariant has been run
func (m *Metrics) RecordInvariantRun(invariant string) {
	m.InvariantRuns.WithLabelValues(invariant).Inc()
}

// RecordAlertGenerated ... Records that an alert has been generated for a given invariant
func (m *Metrics) RecordAlertGenerated(invariant string) {
	m.AlarmsGenerated.WithLabelValues(invariant).Inc()
}

// RecordNodeError ... Records that an error has been caught for a given node
func (m *Metrics) RecordNodeError(node string) {
	m.NodeErrors.WithLabelValues(node).Inc()
}

// Shutdown ... Shuts down the metrics server
func (m *Metrics) Shutdown(ctx context.Context) error {
	return m.server.Shutdown(ctx)
}

// Document ... Returns a list of documented metrics
func (m *Metrics) Document() []DocumentedMetric {
	return m.factory.Document()
}

type noopMetricer struct{}

var NoopMetrics Metricer = new(noopMetricer)

func (n *noopMetricer) IncActiveInvariants()          {}
func (n *noopMetricer) DecActiveInvariants()          {}
func (n *noopMetricer) IncActivePipelines()           {}
func (n *noopMetricer) DecActivePipelines()           {}
func (n *noopMetricer) RecordInvariantRun(_ string)   {}
func (n *noopMetricer) RecordAlertGenerated(_ string) {}
func (n *noopMetricer) RecordNodeError(_ string)      {}
func (n *noopMetricer) RecordUp()                     {}
func (n *noopMetricer) Shutdown(_ context.Context) error {
	return nil
}
func (n *noopMetricer) Start() {}
func (n *noopMetricer) Document() []DocumentedMetric {
	return nil
}
