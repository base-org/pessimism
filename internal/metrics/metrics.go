package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/heuristic"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"go.uber.org/zap"
)

const (
	metricsNamespace    = "pessimism"
	SubsystemHeuristics = "heuristics"
	SubsystemEtl        = "etl"
	batchMethod         = "batch"
)

// serverShutdownTimeout ... Timeout for shutting down the metrics server
const serverShutdownTimeout = 10 * time.Second

// Config ... Metrics server configuration
type Config struct {
	Host              string
	Port              int
	Enabled           bool
	ReadHeaderTimeout int
}

// Metricer ... Interface for metrics
type Metricer interface {
	IncMissedBlock(PathType core.PathID)
	IncActiveHeuristics(ht core.HeuristicType, network core.Network, PathType core.PathType)
	IncActivePipelines(PathType core.PathType, network core.Network)
	DecActivePipelines(PathType core.PathType, network core.Network)
	RecordBlockLatency(network core.Network, latency float64)
	RecordHeuristicRun(heuristic heuristic.Heuristic)
	RecordAlertGenerated(alert core.Alert, dest core.AlertDestination, clientName string)
	RecordNodeError(network core.Network)
	RecordPathLatency(id core.PathID, latency float64)
	RecordAssessmentError(h heuristic.Heuristic)
	RecordInvExecutionTime(h heuristic.Heuristic, latency float64)
	RecordUp()
	Start()
	Shutdown(ctx context.Context) error
	RecordRPCClientRequest(method string) func(err error)
	RecordRPCClientBatchRequest(b []rpc.BatchElem) func(err error)
	Document() []DocumentedMetric
}

// Metrics ... Metrics struct
type Metrics struct {
	rpcClientRequestsTotal          *prometheus.CounterVec
	rpcClientRequestDurationSeconds *prometheus.HistogramVec
	Up                              prometheus.Gauge
	ActivePipelines                 *prometheus.GaugeVec
	ActiveHeuristics                *prometheus.GaugeVec
	HeuristicRuns                   *prometheus.CounterVec
	AlertsGenerated                 *prometheus.CounterVec
	NodeErrors                      *prometheus.CounterVec
	MissedBlocks                    *prometheus.CounterVec
	BlockLatency                    *prometheus.GaugeVec
	PipelineLatency                 *prometheus.GaugeVec
	InvExecutionTime                *prometheus.GaugeVec
	HeuristicErrors                 *prometheus.CounterVec

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
		rpcClientRequestsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Subsystem: SubsystemEtl,
			Name:      "requests_total",
			Help:      "Total RPC requests initiated by the RPC client",
		}, []string{
			"method",
		}),
		rpcClientRequestDurationSeconds: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: metricsNamespace,
			Subsystem: SubsystemEtl,
			Name:      "request_duration_seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			Help:      "Histogram of RPC client request durations",
		}, []string{
			"method",
		}),
		Up: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: metricsNamespace,
			Name:      "up",
			Help:      "1 if the service is up",
		}),
		ActiveHeuristics: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "active_heuristics",
			Help:      "Number of active heuristics",
			Namespace: metricsNamespace,
			Subsystem: SubsystemHeuristics,
		}, []string{"heuristic", "network", "pipeline"}),

		ActivePipelines: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "active_pipelines",
			Help:      "Number of active pipelines",
			Namespace: metricsNamespace,
			Subsystem: SubsystemEtl,
		}, []string{"pipeline", "network"}),

		HeuristicRuns: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "heuristic_runs_total",
			Help:      "Number of times a specific heuristic has been run",
			Namespace: metricsNamespace,
			Subsystem: SubsystemHeuristics,
		}, []string{"network", "heuristic"}),

		AlertsGenerated: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "alerts_generated_total",
			Help:      "Number of total alerts generated for a given heuristic",
			Namespace: metricsNamespace,
		}, []string{"network", "heuristic", "pipeline", "severity", "destination", "client_name"}),

		NodeErrors: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "node_errors_total",
			Help:      "Number of node errors caught",
			Namespace: metricsNamespace,
		}, []string{"node"}),
		BlockLatency: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "block_latency",
			Help:      "Millisecond latency of block processing",
			Namespace: metricsNamespace,
		}, []string{"network"}),

		PipelineLatency: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "pipeline_latency",
			Help:      "Millisecond latency of pipeline processing",
			Namespace: metricsNamespace,
		}, []string{"PathType"}),
		InvExecutionTime: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "heuristic_execution_time",
			Help:      "Nanosecond time of heuristic execution",
			Namespace: metricsNamespace,
		}, []string{"heuristic"}),
		HeuristicErrors: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "heuristic_errors_total",
			Help:      "Number of errors generated by heuristic executions",
			Namespace: metricsNamespace,
		}, []string{"heuristic"}),
		MissedBlocks: factory.NewCounterVec(prometheus.CounterOpts{
			Name:      "missed_blocks_total",
			Help:      "Number of missed blocks",
			Namespace: metricsNamespace,
		}, []string{"PathType"}),

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

// RecordAssessmentError ... Increments the number of errors generated by heuristic executions
func (m *Metrics) RecordAssessmentError(h heuristic.Heuristic) {
	ht := h.Type().String()
	m.HeuristicErrors.WithLabelValues(ht).Inc()
}

// RecordInvExecutionTime ... Records the time it took to execute a heuristic
func (m *Metrics) RecordInvExecutionTime(h heuristic.Heuristic, latency float64) {
	// ht := h.SUUID().PID.HeuristicType().String()
	// m.InvExecutionTime.WithLabelValues(ht).Set(latency)
}

// IncMissedBlock ... Increments the number of missed blocks
func (m *Metrics) IncMissedBlock(id core.PathID) {
	m.MissedBlocks.WithLabelValues(id.String()).Inc()
}

// IncActiveHeuristics ... Increments the number of active heuristics
func (m *Metrics) IncActiveHeuristics(ht core.HeuristicType, n core.Network,
	pt core.PathType) {
	m.ActiveHeuristics.WithLabelValues(ht.String(), n.String(), pt.String()).Inc()
}

// IncActivePipelines ... Increments the number of active pipelines
func (m *Metrics) IncActivePipelines(pt core.PathType, n core.Network) {
	m.ActivePipelines.WithLabelValues(pt.String(), n.String()).Inc()
}

// DecActivePipelines ... Decrements the number of active pipelines
func (m *Metrics) DecActivePipelines(pt core.PathType, n core.Network) {
	m.ActivePipelines.WithLabelValues(pt.String(), n.String()).Dec()
}

// RecordHeuristicRun ... Records that a given heuristic has been run
func (m *Metrics) RecordHeuristicRun(h heuristic.Heuristic) {
	// net := h.SUUID().PID.Network().String()
	// ht := h.SUUID().PID.HeuristicType().String()
	// m.HeuristicRuns.WithLabelValues(net, ht).Inc()
}

// RecordAlertGenerated ... Records that an alert has been generated for a given heuristic
func (m *Metrics) RecordAlertGenerated(alert core.Alert, dest core.AlertDestination, clientName string) {
	net := alert.PathID.Network().String()
	h := alert.HT.String()
	sev := alert.Sev.String()

	m.AlertsGenerated.WithLabelValues(net, h, sev, dest.String(), clientName).Inc()
}

// RecordNodeError ... Records that an error has been caught for a given node
func (m *Metrics) RecordNodeError(n core.Network) {
	m.NodeErrors.WithLabelValues(n.String()).Inc()
}

// RecordBlockLatency ... Records the latency of block processing
func (m *Metrics) RecordBlockLatency(n core.Network, latency float64) {
	m.BlockLatency.WithLabelValues(n.String()).Set(latency)
}

// RecordPathLatency ... Records the latency of pipeline processing
func (m *Metrics) RecordPathLatency(id core.PathID, latency float64) {
	m.PipelineLatency.WithLabelValues(id.String()).Set(latency)
}

func (m *Metrics) RecordRPCClientRequest(method string) func(err error) {
	m.rpcClientRequestsTotal.WithLabelValues(method).Inc()
	// timer := prometheus.NewTimer(m.rpcClientRequestDurationSeconds.WithLabelValues(method))
	// return func(err error) {
	// 	m.recordRPCClientResponse(method, err)
	// 	timer.ObserveDuration()
	// }

	return nil
}

func (m *Metrics) RecordRPCClientBatchRequest(b []rpc.BatchElem) func(err error) {
	m.rpcClientRequestsTotal.WithLabelValues(batchMethod).Add(float64(len(b)))
	for _, elem := range b {
		m.rpcClientRequestsTotal.WithLabelValues(elem.Method).Inc()
	}

	// timer := prometheus.NewTimer(m.rpcClientRequestDurationSeconds.WithLabelValues(batchMethod))
	// return func(err error) {
	// 	m.recordRPCClientResponse(batchMethod, err)
	// 	timer.ObserveDuration()

	// 	// Record errors for individual requests
	// 	for _, elem := range b {
	// 		m.recordRPCClientResponse(elem.Method, elem.Error)
	// 	}
	// }
	return nil
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

func (n *noopMetricer) IncMissedBlock(_ core.PathID) {}
func (n *noopMetricer) RecordUp()                    {}
func (n *noopMetricer) IncActiveHeuristics(_ core.HeuristicType, _ core.Network, _ core.PathType) {
}
func (n *noopMetricer) RecordInvExecutionTime(_ heuristic.Heuristic, _ float64)              {}
func (n *noopMetricer) IncActivePipelines(_ core.PathType, _ core.Network)                   {}
func (n *noopMetricer) DecActivePipelines(_ core.PathType, _ core.Network)                   {}
func (n *noopMetricer) RecordHeuristicRun(_ heuristic.Heuristic)                             {}
func (n *noopMetricer) RecordAlertGenerated(_ core.Alert, _ core.AlertDestination, _ string) {}
func (n *noopMetricer) RecordNodeError(_ core.Network)                                       {}
func (n *noopMetricer) RecordBlockLatency(_ core.Network, _ float64)                         {}
func (n *noopMetricer) RecordPathLatency(_ core.PathID, _ float64)                           {}
func (n *noopMetricer) RecordAssessmentError(_ heuristic.Heuristic)                          {}
func (n *noopMetricer) RecordRPCClientRequest(_ string) func(err error) {
	return func(err error) {}
}
func (n *noopMetricer) RecordRPCClientBatchRequest(_ []rpc.BatchElem) func(err error) {
	return func(err error) {}
}

func (n *noopMetricer) Shutdown(_ context.Context) error {
	return nil
}
func (n *noopMetricer) Start() {}
func (n *noopMetricer) Document() []DocumentedMetric {
	return nil
}
