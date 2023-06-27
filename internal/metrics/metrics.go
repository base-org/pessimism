package metrics

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-go/v5/statsd"
	"strings"
)

const (
	Namespace  = "pessimism"
	MetricsKey = "statsd"
)

type Config struct {
	Host          string
	Port          int
	Environment   string
	EnableMetrics bool
}

var Metrics statsd.ClientInterface
var metrics = Metrics

// NewStatsd ... Initializer for statsd client, returns NoopClient if metrics are disabled
func NewStatsd(cfg *Config) {
	if !cfg.EnableMetrics {
		metrics = &statsd.NoOpClient{}
		return
	}

	host := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var err error
	metrics, err = statsd.New(host, statsd.WithTags([]string{
		fmt.Sprintf("env:%s", cfg.Environment),
		fmt.Sprintf("service:%s", Namespace),
	}))
	if err != nil {
		panic("could not initialize statsd client")
	}
}

func WithContext(ctx context.Context) statsd.ClientInterface {
	if ctx == nil {
		return metrics
	}

	if ctxStats, ok := ctx.Value(MetricsKey).(statsd.ClientInterface); ok {
		return ctxStats
	}

	return metrics

}

func MetricName(scope string, metric string, additionals ...string) string {
	return strings.Join(append([]string{Namespace, scope, metric}, additionals...), ".")
}

// TODO: Impl GetMetadataTags to cast invariant and pipeline configs to metric tags
func GetMetadataTags() []string {

	return []string{}
}
