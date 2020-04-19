module github.com/moorara/observe

go 1.14

require (
	github.com/prometheus/client_golang v1.5.1
	github.com/stretchr/testify v1.5.1
	go.opentelemetry.io/otel v0.4.2
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.4.2
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.4.2
	go.uber.org/zap v1.14.1
)
