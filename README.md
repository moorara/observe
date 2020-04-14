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

  // Initialize the singleton logger
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

Output:

```json
{"level":"info","timestamp":"2020-04-14T10:57:00.63421-04:00","caller":"example/main.go:21","message":"hello, world!","environment":"production","logger":"my-service","region":"us-east-1"}
{"level":"info","timestamp":"2020-04-14T10:57:00.63432-04:00","caller":"example/main.go:24","message":"starting server ...","environment":"production","logger":"my-service","region":"us-east-1","port":8080,"transport":"http"}
{"level":"info","timestamp":"2020-04-14T10:57:00.634331-04:00","caller":"example/main.go:31","message":"starting server ...","environment":"production","logger":"my-service","region":"us-east-1","port":8080,"transport":"grpc"}
```

You can read the full documentation for zap logger [here](https://pkg.go.dev/go.uber.org/zap).


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
