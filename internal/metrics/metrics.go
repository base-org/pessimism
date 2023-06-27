package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"strconv"
)

const metricsNamespace = "pessimism"

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
}

var _ Metricer = (*Metrics)(nil)

func NewMetrics() *Metrics {
	return &Metrics{
		ActiveInvariants: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "active_invariants",
			Help:      "Number of active invariants",
			Namespace: metricsNamespace,
		}),

		ActivePipelines: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "active_pipelines",
			Help:      "Number of active pipelines",
			Namespace: metricsNamespace,
		}),

		InvariantRuns: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "invariant_runs",
			Help:      "Number of times a specific invariant has been run",
			Namespace: metricsNamespace,
		}, []string{"invariant"}),

		AlarmsGenerated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "alarms_generated",
			Help:      "Number of total alarms generated for a given invariant",
			Namespace: metricsNamespace,
		}, []string{"invariant"}),

		NodeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "node_errors",
			Help:      "Number of node errors caught",
			Namespace: metricsNamespace,
		}, []string{"node"}),
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

func (m *Metrics) Serve(hostname string, port uint64) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    net.JoinHostPort(hostname, strconv.FormatUint(port, 10)),
		Handler: mux,
	}

	err := srv.ListenAndServe()

	return srv, err
}

type noopMetricer struct{}

var NoopMetrics Metricer = new(noopMetricer)

func (n *noopMetricer) IncActiveInvariants()                  {}
func (n *noopMetricer) DecActiveInvariants()                  {}
func (n *noopMetricer) IncActivePipelines()                   {}
func (n *noopMetricer) DecActivePipelines()                   {}
func (n *noopMetricer) RecordInvariantRun(invariant string)   {}
func (n *noopMetricer) RecordAlarmGenerated(invariant string) {}
func (n *noopMetricer) RecordNodeError(node string)           {}
