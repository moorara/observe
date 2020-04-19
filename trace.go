package observe

import (
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/key"
	api "go.opentelemetry.io/otel/api/trace"
	exporter "go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdk "go.opentelemetry.io/otel/sdk/trace"
)

// TracerOptions are optional configurations for creating a tracer.
type TracerOptions struct {
	Name              string
	Tags              [][2]string
	AgentEndpoint     string
	CollectorEndpoint string
	CollectorUserName string
	CollectorPassword string
}

// NewTracer creates a new OpenTelemetry Tracer.
func NewTracer(opts TracerOptions) (api.Tracer, func()) {
	var endpointOpt exporter.EndpointOption
	switch {
	case opts.AgentEndpoint != "":
		endpointOpt = exporter.WithAgentEndpoint(opts.AgentEndpoint)
	case opts.CollectorEndpoint != "":
		endpointOpt = exporter.WithCollectorEndpoint(
			opts.CollectorEndpoint,
			exporter.WithUsername(opts.CollectorUserName),
			exporter.WithPassword(opts.CollectorPassword),
		)
	}

	tags := []core.KeyValue{}
	for _, pair := range opts.Tags {
		tags = append(tags, key.String(pair[0], pair[1]))
	}

	processOpt := exporter.WithProcess(
		exporter.Process{
			ServiceName: opts.Name,
			Tags:        tags,
		},
	)

	sdkOpt := exporter.WithSDK(
		&sdk.Config{
			DefaultSampler: sdk.AlwaysSample(),
		},
	)

	provider, close, err := exporter.NewExportPipeline(endpointOpt, processOpt, sdkOpt)
	if err != nil {
		panic(err)
	}

	global.SetTraceProvider(provider)
	tracer := global.TraceProvider().Tracer(opts.Name)

	return tracer, close
}
