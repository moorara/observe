[![Go Doc][godoc-image]][godoc-url]
[![Build Status][workflow-image]][workflow-url]
[![Go Report Card][goreport-image]][goreport-url]
[![Test Coverage][coverage-image]][coverage-url]
[![Maintainability][maintainability-image]][maintainability-url]

# observe

This repo provides a single package for building observable Go applications.
It leverages the [OpenTelemetry](https://opentelemetry.io) API in an opinionated way.
This package aims to unify three pillars of observability in one single package that is _easy-to-use_ and _hard-to-misuse_.

Sub-packages can be used for implementing observability for different transport protocols out of the box.

## Logging

Logs are used for _auditing_ purposes (sometimes for debugging with limited capabilities).
When looking at logs, you need to know what to look for ahead of the time (known unknowns vs. unknown unknowns).
Since log data can have any arbitrary shape and size, they cannot be used for real-time computational purposes.
Logs are hard to track across different and distributed processes. Logs are also very expensive at scale.

### Quick Start

<details>
  <summary>Logging example</summary>

```go
package main

import (
  "github.com/moorara/observe"
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"
)

func main() {
  logger, config := observe.NewZapLogger(observe.LoggerOptions{
    Name:        "my-service",
    Environment: "production",
    Region:      "us-east-1",
    Level:       "info",
  })

  // Initializing the singleton logger
  observe.SetLogger(logger)

  // Logging using the singleton logger
  observe.Logger.Info("hello, world!")

  // Logging using the typed logger
  logger.Info("starting server ...",
    zap.Int("port", 8080),
    zap.String("transport", "http"),
  )

  // Logging using the untyped logger
  sugared := logger.Sugar()
  sugared.Infow("starting server ...",
    "port", 8080,
    "transport", "grpc",
  )

  // Changing the logging level
  config.Level.SetLevel(zapcore.WarnLevel)

  // These logs will not get printed
  observe.Logger.Info("request received")
  logger.Info("request processed")
  sugared.Info("response sent")
}
```

Here is the outpiut from _stdout_:

```json
{"level":"info","timestamp":"2020-04-14T10:57:00.63421-04:00","caller":"example/main.go:21","message":"hello, world!","environment":"production","logger":"my-service","region":"us-east-1"}
{"level":"info","timestamp":"2020-04-14T10:57:00.63432-04:00","caller":"example/main.go:24","message":"starting server ...","environment":"production","logger":"my-service","region":"us-east-1","port":8080,"transport":"http"}
{"level":"info","timestamp":"2020-04-14T10:57:00.634331-04:00","caller":"example/main.go:31","message":"starting server ...","environment":"production","logger":"my-service","region":"us-east-1","port":8080,"transport":"grpc"}
```
</details>

### Documentation

  - [go.uber.org/zap](https://pkg.go.dev/go.uber.org/zap)

## Metrics

Metrics are _regular time-series_ data with _low and fixed cardinality_.
They are aggregated by time. Metrics are used for **real-time** monitoring purposes.
Using metrics with can implement **SLIs** (service-level indicators), **SLOs** (service-level objectives), and automated alerting.
Metrics are very good at taking the distribution of data into account.
Metrics cannot be used with _high-cardinality data_.

### Quick Start

Metric instruments capture measurements. A Meter is used for creating metric instruments.

Three kind of metric instruments:

  - **Counter**:  metric events that _Add_ to a value that is summed over time.
  - **Measure**:  metric events that _Record_ a value that is aggregated over time.
  - **Observer**: metric events that _Observe_ a coherent set of values at a point in time.

Counter and Measure instruments use synchronous APIs for capturing measurements driven by events in the application.
These measurements are associated with a distributed context (_correlation context_).

Observer instruments use an asynchronous API (callback) for collecting measurements on intervals.
They are used to report measurements about the state of the application periodically.
Observer instruments do not have a distributed context (_correlation context_) since they are reported outside of a context.

Aggregation is the process of combining a large number of measurements into exact or estimated statistics.
The type of aggregation is determined by the metric instrument implementation.

  - Counter instruments use _Sum_ aggregation
  - Measure instruments use _MinMaxSumCount_ aggregation
  - Observer instruments use _LastValue_ aggregation.

The Metric SDK specification allow configuring alternative aggregations for metric instruments.

<details>
  <summary>Metrics example</summary>

```go
package main

import (
  "context"
  "log"
  "net/http"
  "runtime"
  "time"

  "github.com/moorara/observe"
  "go.opentelemetry.io/otel/api/core"
  "go.opentelemetry.io/otel/api/correlation"
  "go.opentelemetry.io/otel/api/key"
  "go.opentelemetry.io/otel/api/metric"
)

func callback(result metric.Int64ObserverResult) {
  ms := new(runtime.MemStats)
  runtime.ReadMemStats(ms)
  result.Observe(int64(ms.Alloc),
    key.String("function", "ReadMemStats"),
  )
}

type instruments struct {
  meter        metric.Meter
  reqCounter   metric.Int64Counter
  reqDuration  metric.Float64Measure
  allocatedMem metric.Int64Observer
}

func newInstruments(meter metric.Meter) *instruments {
  mustMeter := metric.Must(meter)

  return &instruments{
    meter:        meter,
    reqCounter:   mustMeter.NewInt64Counter("requests_total", metric.WithDescription("the total number of requests")),
    reqDuration:  mustMeter.NewFloat64Measure("request_duration_seconds", metric.WithDescription("the duration of requests in seconds")),
    allocatedMem: mustMeter.RegisterInt64Observer("allocated_memory_bytes", callback, metric.WithDescription("number of bytes allocated and in use")),
  }
}

func (i *instruments) counterExample(ctx context.Context) {
  bounded := i.reqCounter.Bind(
    key.String("protocol", "http"),
  )
  bounded.Add(ctx, 1)
}

func (i *instruments) measureExample(ctx context.Context) {
  bounded := i.reqDuration.Bind(
    key.Bool("success", true),
  )
  d := 100 * time.Millisecond
  bounded.Record(ctx, d.Seconds())
}

func (i *instruments) recordExample(ctx context.Context) {
  labels := []core.KeyValue{
    key.Uint("status", 200),
  }
  d := 200 * time.Millisecond
  i.meter.RecordBatch(ctx, labels,
    i.reqCounter.Measurement(1),
    i.reqDuration.Measurement(d.Seconds()),
  )
}

func main() {
  meter, close, handler := observe.NewMeter(observe.MeterOptions{
    Name: "my_service",
  })

  // Flushing all metrics before exiting
  defer close()

  i := newInstruments(meter)

  // Creating a correlation context
  ctx := context.Background()
  ctx = correlation.NewContext(ctx,
    key.String("topic", "greeting"),
  )

  i.counterExample(ctx)
  i.measureExample(ctx)
  i.recordExample(ctx)

  // Serving metrics endpoint
  http.Handle("/metrics", handler)
  log.Fatal(http.ListenAndServe(":8080", nil))
}

```

Here is the output from _http://localhost:8080/metrics_:

```
# HELP allocated_memory_bytes number of bytes allocated and in use
# TYPE allocated_memory_bytes summary
allocated_memory_bytes{quantile="0.1"} 373928
allocated_memory_bytes{quantile="0.5"} 3.138576e+06
allocated_memory_bytes{quantile="0.95"} 3.823808e+06
allocated_memory_bytes{quantile="0.99"} 3.823808e+06
allocated_memory_bytes_sum 2.6009896e+07
allocated_memory_bytes_count 10
# HELP request_duration_seconds the duration of requests in seconds
# TYPE request_duration_seconds summary
request_duration_seconds{quantile="0.1"} 0.2
request_duration_seconds{quantile="0.5"} 0.2
request_duration_seconds{quantile="0.95"} 0.2
request_duration_seconds{quantile="0.99"} 0.2
request_duration_seconds_sum 0.30000000000000004
request_duration_seconds_count 2
# HELP requests_total the total number of requests
# TYPE requests_total counter
requests_total 2
```
</details>

### Documentation

  - [Metrics API](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/metrics/api.md)
  - [Metric User-Facing API](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/metrics/api-user.md)
  - [go.opentelemetry.io/otel/api/metric](https://pkg.go.dev/go.opentelemetry.io/otel/api/metric)

## Tracing

Traces are used for _debugging_ and _tracking_ requests across different processes and services.
They can be used for identifying performance bottlenecks.
Due to their very data-heavy nature, traces in real-world applications need to be _sampled_.
Insights extracted from traces cannot be aggregated since they are sampled.
In other words, information captured by one trace does not tell anything about how this trace is compared against other traces and what is the distribution of data.

### Quick Start

<details>
  <summary>Tracing example</summary>

```go
package main

import (
  "context"
  "time"

  "github.com/moorara/observe"
  "go.opentelemetry.io/otel/api/correlation"
  "go.opentelemetry.io/otel/api/key"
  "go.opentelemetry.io/otel/api/trace"
)

type handler struct {
  tracer trace.Tracer
}

func newHandler(tracer trace.Tracer) *handler {
  return &handler{
    tracer: tracer,
  }
}

func (h *handler) handle(ctx context.Context) {
  childCtx, span := h.tracer.Start(ctx, "request")
  defer span.End()

  h.fetch(childCtx)
  h.respond(childCtx)
}

func (h *handler) fetch(ctx context.Context) {
  _, span := h.tracer.Start(ctx, "read-database")
  defer span.End()

  time.Sleep(100 * time.Millisecond)
}

func (h *handler) respond(ctx context.Context) {
  _, span := h.tracer.Start(ctx, "send-response")
  defer span.End()

  time.Sleep(10 * time.Millisecond)
}

func main() {
  tracer, close := observe.NewTracer(observe.TracerOptions{
    Name: "my-service",
    Tags: [][2]string{
      {"version", "0.1.0"},
    },
    AgentEndpoint: "localhost:6831",
  })

  // Flushing all traces before exiting
  defer close()

  h := newHandler(tracer)

  // Creating a correlation context
  ctx := context.Background()
  ctx = correlation.NewContext(ctx,
    key.String("topic", "greeting"),
  )

  h.handle(ctx)
}
```
</details>

### Documentation

  - [Tracing API](https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/api.md)
  - [go.opentelemetry.io/otel/api/trace](https://pkg.go.dev/go.opentelemetry.io/otel/api/trace)


[godoc-url]: https://pkg.go.dev/github.com/moorara/observe
[godoc-image]: https://godoc.org/github.com/moorara/observe?status.svg
[workflow-url]: https://github.com/moorara/observe/actions
[workflow-image]: https://github.com/moorara/observe/workflows/Main/badge.svg
[goreport-url]: https://goreportcard.com/report/github.com/moorara/observe
[goreport-image]: https://goreportcard.com/badge/github.com/moorara/observe
[coverage-url]: https://codeclimate.com/github/moorara/observe/test_coverage
[coverage-image]: https://api.codeclimate.com/v1/badges/ae0da137cc52c257a27a/test_coverage
[maintainability-url]: https://codeclimate.com/github/moorara/observe/maintainability
[maintainability-image]: https://api.codeclimate.com/v1/badges/ae0da137cc52c257a27a/maintainability
