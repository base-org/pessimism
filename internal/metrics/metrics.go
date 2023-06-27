package metrics

import (
	"context"
	"github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "pessimism"

const (
	SubsystemInvariants = "invariants"
	SubsystemEtl        = "etl"
)

type Config struct {
	Host          string
	Port          uint64
	EnableMetrics bool
}

type Metricer interface {
	IncActiveInvariants()
	DecActiveInvariants()
	IncActivePipelines()
	DecActivePipelines()
	RecordInvariantRun(invariant string)
	RecordAlarmGenerated(invariant string)
	RecordNodeError(node string)
}

type Metrics struct {
	ActiveInvariants prometheus.Gauge
	ActivePipelines  prometheus.Gauge
	InvariantRuns    *prometheus.CounterVec
	AlarmsGenerated  *prometheus.CounterVec
	NodeErrors       *prometheus.CounterVec

	registry *prometheus.Registry
	factory  metrics.Factory
}

var _ Metricer = (*Metrics)(nil)

func NewMetrics() *Metrics {

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	factory := metrics.With(registry)

	return &Metrics{
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
	}
}

func (m *Metrics) IncActiveInvariants() {
	m.ActiveInvariants.Inc()
}

func (m *Metrics) DecActiveInvariants() {
	m.ActiveInvariants.Dec()
}

func (m *Metrics) IncActivePipelines() {
	m.ActivePipelines.Inc()
}

func (m *Metrics) DecActivePipelines() {
	m.ActivePipelines.Dec()
}

func (m *Metrics) RecordInvariantRun(invariant string) {
	m.InvariantRuns.WithLabelValues(invariant).Inc()
}

func (m *Metrics) RecordAlarmGenerated(invariant string) {
	m.AlarmsGenerated.WithLabelValues(invariant).Inc()
}

func (m *Metrics) RecordNodeError(node string) {
	m.NodeErrors.WithLabelValues(node).Inc()
}

func (m *Metrics) Serve(ctx context.Context, cfg *Config) error {
	addr := net.JoinHostPort(cfg.Host, strconv.FormatUint(cfg.Port, 10))
	server := &http.Server{}
	server.Handler = promhttp.InstrumentMetricHandler(m.registry, promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
	server.Addr = addr

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	return server.ListenAndServe()
}

func (m *Metrics) Document() []metrics.DocumentedMetric {
	return m.factory.Document()
}

type noopMetricer struct{}

var NoopMetrics Metricer = new(noopMetricer)

func (n *noopMetricer) IncActiveInvariants()          {}
func (n *noopMetricer) DecActiveInvariants()          {}
func (n *noopMetricer) IncActivePipelines()           {}
func (n *noopMetricer) DecActivePipelines()           {}
func (n *noopMetricer) RecordInvariantRun(_ string)   {}
func (n *noopMetricer) RecordAlarmGenerated(_ string) {}
func (n *noopMetricer) RecordNodeError(_ string)      {}
