# observe

This repo includes core libraries for implementing observability in Go applications.
Observability is comprised of the following pillars:

  - Logging
  - Metrics
  - Tracing
  - Reporting (Events)

## log

This package can be used for implementing **structured logging** in Go applications.
It supports four different logging levels and `JSON` logging format.

Logs are used for _auditing_ purposes (sometimes for debugging with limited capabilities).
Since log data can have any arbitrary shape and size, they cannot be used for real-time computational purposes.
Logs are also hard to track across different and distributed processes.

## metrics

This package can be used for implementing **metrics** in Go applications.
It supports [OpenMetrics](https://openmetrics.io) format and uses [Prometheus](https://prometheus.io) API.

Metrics are _regular time-series_ data with _low and fixed cardinality_.
Metrics are used for defining **SLIs** (service-level indicators), **SLOs** (service-level objectives), and automated alerting.
Metrics are very good at taking the distribution of data into account.
Metrics cannot be used with _high-cardinality data_.

## trace

This package can be used for implementing **tracing** in Go applications.
It supports [OpenTracing](https://opentracing.io/) format and uses [Jaeger](https://www.jaegertracing.io).

Traces are used for _debugging_ and _tracking_ requests across different processes and services.
They can also be used for identifying performance bottlenecks.
Due to their very data-heavy nature, traces in real-world applications need to be _sampled_.
Insights extracted from traces cannot be aggregated due to the fact that they are sampled.
In other words, information captured by one trace do not tell anything about how this trace is compared against other traces and what is the distribution of data.

## report

The package can be used for implementing **error and event reporting** in Go applications.
It uses [Rollbar](https://rollbar.com) API.

Events are also _irregular time-series_ data and can have arbitrary number of metadata.
They occur in temporal order, but the interval between occurrences are inconsistent and sporadic.
Reporting is used for alerting on important or critical events such as errors, crashes, etc.
