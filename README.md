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

Metrics are _time-series_ data with _low and fixed cardinality_.
Metrics are used for defining **SLIs** (service-level indicators), **SLOs** (service-level objectives), and automated alerting.
Metrics cannot be used with _high-cardinality data_.
