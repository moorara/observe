package main

import (
	"net/http"

	"github.com/moorara/observe/log"
	"github.com/moorara/observe/metrics"
	"github.com/moorara/observe/trace"
	xhttp "github.com/moorara/observe/xhttp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const port = ":10080"

func main() {
	// Create a logger
	logger := log.NewLogger(log.Options{
		Name:        "server",
		Environment: "dev",
		Region:      "us-east-1",
	})

	// Create a metrics factory
	mf := metrics.NewFactory(metrics.FactoryOptions{})

	// Create a tracer
	tracer, closer, _ := trace.NewTracer(trace.Options{Name: "server"})
	defer closer.Close()

	// Create an http server middleware
	mid := xhttp.NewServerMiddleware(
		xhttp.ServerLogging(logger),
		xhttp.ServerMetrics(mf),
		xhttp.ServerTracing(tracer),
	)

	s := &server{tracer: tracer}
	h := mid.Metrics(mid.RequestID(mid.Tracing(mid.Logging(s.handler))))

	http.Handle("/", h)
	http.Handle("/metrics", promhttp.Handler())
	logger.Infof("starting http server on port %s ...", port)
	panic(http.ListenAndServe(port, nil))
}
