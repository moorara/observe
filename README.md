[![Go Doc][godoc-image]][godoc-url]
[![Build Status][workflow-image]][workflow-url]
[![Go Report Card][goreport-image]][goreport-url]
[![Test Coverage][coverage-image]][coverage-url]
[![Maintainability][maintainability-image]][maintainability-url]

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
When looking at logs, you need to know what to look for ahead of the time (known unknowns vs. unknown unknowns).
Since log data can have any arbitrary shape and size, they cannot be used for real-time computational purposes.
Logs are hard to track across different and distributed processes. Logs are also very expensive at scale.

## metrics

This package can be used for implementing **metrics** in Go applications.
It supports [OpenMetrics](https://openmetrics.io) format and uses [Prometheus](https://prometheus.io) API.

Metrics are _regular time-series_ data with _low and fixed cardinality_.
They are aggregated by time. Metrics are used for **real-time** monitoring purposes.
Using metrics with can implement **SLIs** (service-level indicators), **SLOs** (service-level objectives), and automated alerting.
Metrics are very good at taking the distribution of data into account.
Metrics cannot be used with _high-cardinality data_.

## trace

This package can be used for implementing **tracing** in Go applications.
It supports [OpenTracing](https://opentracing.io/) format and uses [Jaeger](https://www.jaegertracing.io).

Traces are used for _debugging_ and _tracking_ requests across different processes and services.
They can be used for identifying performance bottlenecks.
Due to their very data-heavy nature, traces in real-world applications need to be _sampled_.
Insights extracted from traces cannot be aggregated since they are sampled.
In other words, information captured by one trace does not tell anything about how this trace is compared against other traces and what is the distribution of data.

## report

The package can be used for implementing **error and event reporting** in Go applications.
It uses [Rollbar](https://rollbar.com) API.

Events are _irregular time-series_ data and can have an arbitrary number of metadata.
They occur in temporal order, but the interval between occurrences is inconsistent and sporadic.
Events are used for reporting and alerting on important or critical events such as errors, crashes, etc.


[godoc-url]: https://godoc.org/github.com/moorara/observe
[godoc-image]: https://godoc.org/github.com/moorara/observe?status.svg
[workflow-url]: https://github.com/moorara/observe/actions
[workflow-image]: https://github.com/moorara/observe/workflows/Main/badge.svg
[goreport-url]: https://goreportcard.com/report/github.com/moorara/observe
[goreport-image]: https://goreportcard.com/badge/github.com/moorara/observe
[coverage-url]: https://codeclimate.com/github/moorara/observe/test_coverage
[coverage-image]: https://api.codeclimate.com/v1/badges/ae0da137cc52c257a27a/test_coverage
[maintainability-url]: https://codeclimate.com/github/moorara/observe/maintainability
[maintainability-image]: https://api.codeclimate.com/v1/badges/ae0da137cc52c257a27a/maintainability
