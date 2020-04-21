package observe

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/api/global"
	api "go.opentelemetry.io/otel/api/metric"
	exporter "go.opentelemetry.io/otel/exporters/metric/prometheus"
)

const (
	defaultInterval = 5 * time.Second
)

var (
	defaultBuckets   = []float64{0.01, 0.10, 0.50, 1.00, 5.00}
	defaultQuantiles = []float64{0.1, 0.5, 0.95, 0.99}
)

// MeterOptions are optional configurations for creating a meter.
type MeterOptions struct {
	Name          string
	Buckets       []float64
	Quantiles     []float64
	SystemMetrics bool
	Interval      time.Duration
}

// NewMeter creates a new OpenTelemetry Meter.
func NewMeter(opts MeterOptions) (api.Meter, func(), http.Handler) {
	if len(opts.Buckets) == 0 {
		opts.Buckets = defaultBuckets
	}

	if len(opts.Quantiles) == 0 {
		opts.Quantiles = defaultQuantiles
	}

	if opts.Interval == 0 {
		opts.Interval = defaultInterval
	}

	// Create a new Prometheus registry
	registry := prometheus.NewRegistry()
	if opts.SystemMetrics {
		registry.MustRegister(prometheus.NewGoCollector())
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	config := exporter.Config{
		Registerer:              registry,
		Gatherer:                registry,
		DefaultSummaryQuantiles: opts.Quantiles,
		// TODO: Buckets
	}

	controller, handler, err := exporter.NewExportPipeline(config, opts.Interval)
	if err != nil {
		panic(err)
	}

	global.SetMeterProvider(controller)
	meter := global.MeterProvider().Meter(opts.Name)

	return meter, controller.Stop, handler
}
