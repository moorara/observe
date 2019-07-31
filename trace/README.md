# trace

This is a helper package for creating a [Jaeger](https://www.jaegertracing.io) tracer
that creates traces in [OpenTracing](https://opentracing.io) format.

## Quick Start

For creating a tracer with default configurations (*constant sampler* and *agent reporter*):

```go
package main

import (
  "github.com/moorara/observe/trace"
  "github.com/opentracing/opentracing-go/ext"
  "github.com/opentracing/opentracing-go/log"
)

func main() {
  tracer, closer, _ := trace.NewTracer(trace.Options{Name: "service-name"})
  defer closer.Close()

  span := tracer.StartSpan("hello-world")
  defer span.Finish()

  // https://github.com/opentracing/specification/blob/master/semantic_conventions.md
  span.SetTag("protocol", "HTTP")
  ext.HTTPMethod.Set(span, "GET")
  ext.HTTPStatusCode.Set(span, 200)
  span.LogFields(
    log.String("environment", "prodcution"),
    log.String("region", "us-east-1"),
  )
}
```
