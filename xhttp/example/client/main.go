package main

import (
	"net/http"

	"github.com/moorara/observe/log"
	"github.com/moorara/observe/metrics"
	"github.com/moorara/observe/trace"
	xhttp "github.com/moorara/observe/xhttp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const port = ":10081"
const serverAddress = "http://localhost:10080"

func main() {
	// Create a logger
	logger := log.NewLogger(log.Options{
		Name:        "client",
		Environment: "dev",
		Region:      "us-east-1",
	})

	// Create a metrics factory
	mf := metrics.NewFactory(metrics.FactoryOptions{})

	// Create a tracer
	tracer, closer, _ := trace.NewTracer(trace.Options{Name: "client"})
	defer closer.Close()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Infof("starting http server on port %s ...", port)
		panic(http.ListenAndServe(port, nil))
	}()

	// Create an http client middleware
	mid := xhttp.NewClientMiddleware(
		xhttp.ClientLogging(logger),
		xhttp.ClientMetrics(mf),
		xhttp.ClientTracing(tracer),
	)

	c := &client{
		logger: logger,
		mid:    mid,
	}

	for {
		c.call()
	}
}
