package metrics

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/base-org/pessimism/internal/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// spawnServer ... Spawns a new HTTP metrics server
func spawnServer(server *http.Server) {
	logging.NoContext().Info("Starting metrics server",
		zap.String("address", server.Addr))

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		panic(err)
	}
}

// Start ... Starts a listen and serve API routine
func (m *Metrics) Start() {
	go spawnServer(m.server)
}

// initServer ... Initializes a new HTTP server struct object
func initServer(config *Config, registry *prometheus.Registry) *http.Server {
	return &http.Server{
		ReadHeaderTimeout: time.Duration(config.ReadHeaderTimeout) * time.Second,
		Addr:              net.JoinHostPort(config.Host, strconv.FormatUint(uint64(config.Port), 10)),
		Handler: promhttp.InstrumentMetricHandler(
			registry,
			promhttp.HandlerFor(registry, promhttp.HandlerOpts{})),
	}
}
