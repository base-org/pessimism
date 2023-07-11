package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
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
	IncActiveInvariants(invType, network, pipelineType string)
	IncActivePipelines(pipelineType, network string)
	DecActivePipelines(pipelineType, network string)
	RecordInvariantRun(invariant invariant.Invariant)
	RecordAlertGenerated(alert core.Alert)
	RecordNodeError(node string)
	RecordUp()
	Start()
	Shutdown(ctx context.Context) error
	Document() []DocumentedMetric
}

type Metrics struct {
	Up               prometheus.Gauge
	ActivePipelines  *prometheus.GaugeVec
	ActiveInvariants *prometheus.GaugeVec
	InvariantRuns    *prometheus.CounterVec
	AlertsGenerated  *prometheus.CounterVec
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
		ActiveInvariants: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "active_invariants",
			Help:      "Number of active invariants",
			Namespace: metricsNamespace,
			Subsystem: SubsystemInvariants,
		}, []string{"invariant", "network", "pipeline"}),

		ActivePipelines: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "active_pipelines",
			Help:      "Number of active pipelines",
			Namespace: metricsNamespace,
			Subsystem: SubsystemEtl,
		}, []string{"pipeline", "network"}),

		InvariantRuns: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "invariant_runs_total",
			Help:      "Number of times a specific invariant has been run",
			Namespace: metricsNamespace,
			Subsystem: SubsystemInvariants,
		}, []string{"network", "invariant"}),

		AlertsGenerated: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "alerts_generated_total",
			Help:      "Number of total alerts generated for a given invariant",
			Namespace: metricsNamespace,
		}, []string{"network", "invariant", "pipeline", "destination"}),

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
func (m *Metrics) IncActiveInvariants(invType, network, pipelineType string) {
	m.ActiveInvariants.WithLabelValues(invType, network, pipelineType).Inc()
}

// IncActivePipelines ... Increments the number of active pipelines
func (m *Metrics) IncActivePipelines(pipelineType, network string) {
	m.ActivePipelines.WithLabelValues(pipelineType, network).Inc()
}

// DecActivePipelines ... Decrements the number of active pipelines
func (m *Metrics) DecActivePipelines(pipelineType, network string) {
	m.ActivePipelines.WithLabelValues(pipelineType, network).Dec()
}

// RecordInvariantRun ... Records that a given invariant has been run
func (m *Metrics) RecordInvariantRun(inv invariant.Invariant) {
	net := inv.SUUID().PID.Network().String()
	invType := inv.SUUID().PID.InvType().String()
	m.InvariantRuns.WithLabelValues(net, invType).Inc()
}

// RecordAlertGenerated ... Records that an alert has been generated for a given invariant
func (m *Metrics) RecordAlertGenerated(alert core.Alert) {
	net := alert.SUUID.PID.Network().String()
	inv := alert.SUUID.PID.InvType().String()
	pipeline := alert.Ptype.String()
	dest := alert.Dest.String()
	m.AlertsGenerated.WithLabelValues(net, inv, pipeline, dest).Inc()
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

func (n *noopMetricer) IncActiveInvariants(_, _, _ string)       {}
func (n *noopMetricer) DecActiveInvariants(_, _, _ string)       {}
func (n *noopMetricer) IncActivePipelines(_, _ string)           {}
func (n *noopMetricer) DecActivePipelines(_, _ string)           {}
func (n *noopMetricer) RecordInvariantRun(_ invariant.Invariant) {}
func (n *noopMetricer) RecordAlertGenerated(_ core.Alert)        {}
func (n *noopMetricer) RecordNodeError(_ string)                 {}
func (n *noopMetricer) RecordUp()                                {}
func (n *noopMetricer) Shutdown(_ context.Context) error {
	return nil
}
func (n *noopMetricer) Start() {}
func (n *noopMetricer) Document() []DocumentedMetric {
	return nil
}
