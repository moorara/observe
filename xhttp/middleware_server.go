package xhttp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/moorara/observe/log"
	"github.com/moorara/observe/metrics"
	"github.com/moorara/observe/request"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	serverKind                = "server"
	serverSpanName            = "http-server-request"
	serverGaugeMetricName     = "http_server_requests"
	serverCounterMetricName   = "http_server_requests_total"
	serverHistogramMetricName = "http_server_request_duration_seconds"
	serverSummaryMetricName   = "http_server_request_duration_quantiles_seconds"
)

// ContextForTest takes in a request context and inserts a RequestID as well as a new Void Logger.
// For use in tests only, to test functions which expect a logger and RequestID to have been added by the middleware.
func ContextForTest(ctx context.Context) context.Context {
	ctx = request.ContextWithID(ctx, request.NewID())
	ctx = log.ContextWithLogger(ctx, log.NewVoidLogger())
	return ctx
}

// ServerMiddleware is an http server middleware for logging, metrics, tracing, etc.
type ServerMiddleware struct {
	logger  *log.Logger
	metrics *metrics.RequestMetrics
	tracer  opentracing.Tracer
}

// ServerMiddlewareOption sets optional parameters for server middleware.
type ServerMiddlewareOption func(*ServerMiddleware)

// ServerLogging is the option for server middleware to enable logging for every request.
func ServerLogging(logger *log.Logger) ServerMiddlewareOption {
	return func(i *ServerMiddleware) {
		i.logger = logger
	}
}

// ServerMetrics is the option for server middleware to enable metrics for every request.
func ServerMetrics(mf *metrics.Factory) ServerMiddlewareOption {
	metrics := &metrics.RequestMetrics{
		ReqGauge:        mf.Gauge(serverGaugeMetricName, "gauge metric for number of active server-side http requests", []string{"method", "url"}),
		ReqCounter:      mf.Counter(serverCounterMetricName, "counter metric for total number of server-side http requests", []string{"method", "url", "statusCode", "statusClass"}),
		ReqDurationHist: mf.Histogram(serverHistogramMetricName, "histogram metric for duration of server-side http requests in seconds", []string{"method", "url", "statusCode", "statusClass"}),
		ReqDurationSumm: mf.Summary(serverSummaryMetricName, "summary metric for duration of server-side http requests in seconds", []string{"method", "url", "statusCode", "statusClass"}),
	}

	return func(i *ServerMiddleware) {
		i.metrics = metrics
	}
}

// ServerTracing is the option for server middleware to enable tracing for every request.
func ServerTracing(tracer opentracing.Tracer) ServerMiddlewareOption {
	return func(i *ServerMiddleware) {
		i.tracer = tracer
	}
}

// NewServerMiddleware creates a new instance of http server middleware.
func NewServerMiddleware(opts ...ServerMiddlewareOption) *ServerMiddleware {
	sm := &ServerMiddleware{}
	for _, opt := range opts {
		opt(sm)
	}

	return sm
}

// RequestID ensures incoming requests have unique ids.
// This middleware ensures the request headers and context have a unique id.
// A new request id will be generated if needed.
func (m *ServerMiddleware) RequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure request id in headers
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = request.NewID()
			r.Header.Set(requestIDHeader, requestID)
		}

		// Add request id to context
		ctx := r.Context()
		ctx = request.ContextWithID(ctx, requestID)
		req := r.WithContext(ctx)

		// Add request id to response headers
		w.Header().Set(requestIDHeader, requestID)

		// Call the next http handler
		next(w, req)
	}
}

// Logging takes care of logging for incoming http requests.
// Request id will be read from reqeust headers if present.
func (m *ServerMiddleware) Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proto := r.Proto
		method := r.Method
		url := r.URL.Path

		// Create a new logger that logs the context of current request
		logger := m.logger.With(
			"http.kind", serverKind,
			"req.proto", proto,
			"req.method", method,
			"req.url", url,
		)

		if requestID := r.Header.Get(requestIDHeader); requestID != "" {
			logger = logger.With("requestId", requestID)
		}

		// Update request context
		ctx := r.Context()
		ctx = log.ContextWithLogger(ctx, logger)
		req := r.WithContext(ctx)

		// Call the next http handler
		start := time.Now()
		rw := NewResponseWriter(w)
		next(rw, req)
		statusCode := rw.StatusCode
		statusClass := rw.StatusClass
		duration := time.Since(start).Seconds()

		pairs := []interface{}{
			"res.statusCode", statusCode,
			"res.statusClass", statusClass,
			"responseTime", duration,
			"message", fmt.Sprintf("%s %s %d %f", method, url, statusCode, duration),
		}

		// Logging
		switch {
		case statusCode >= 500:
			logger.Error(pairs...)
		case statusCode >= 400:
			logger.Warn(pairs...)
		case statusCode >= 100:
			fallthrough
		default:
			logger.Info(pairs...)
		}
	}
}

// Metrics takes care of metrics for incoming http requests.
func (m *ServerMiddleware) Metrics(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		url := r.URL.Path

		// Increment guage metric
		m.metrics.ReqGauge.WithLabelValues(method, url).Inc()

		// Call the next http handler
		start := time.Now()
		rw := NewResponseWriter(w)
		next(rw, r)
		statusCode := rw.StatusCode
		statusClass := rw.StatusClass
		duration := time.Since(start).Seconds()

		// Metrics
		statusText := strconv.Itoa(statusCode)
		m.metrics.ReqGauge.WithLabelValues(method, url).Dec()
		m.metrics.ReqCounter.WithLabelValues(method, url, statusText, statusClass).Inc()
		m.metrics.ReqDurationHist.WithLabelValues(method, url, statusText, statusClass).Observe(duration)
		m.metrics.ReqDurationSumm.WithLabelValues(method, url, statusText, statusClass).Observe(duration)
	}
}

func (m *ServerMiddleware) createSpan(r *http.Request) opentracing.Span {
	var span opentracing.Span

	carrier := opentracing.HTTPHeadersCarrier(r.Header)
	parentSpanContext, err := m.tracer.Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		span = m.tracer.StartSpan(serverSpanName)
	} else {
		span = m.tracer.StartSpan(serverSpanName, opentracing.ChildOf(parentSpanContext))
	}

	return span
}

// Tracing takes care of tracing for incoming http requests.
// Trace information will be read from reqeust headers if present.
func (m *ServerMiddleware) Tracing(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proto := r.Proto
		method := r.Method
		url := r.URL.Path

		// Create a new span
		span := m.createSpan(r)
		defer span.Finish()

		// Update request context
		ctx := r.Context()
		ctx = opentracing.ContextWithSpan(ctx, span)
		req := r.WithContext(ctx)

		// Call the next http handler
		rw := NewResponseWriter(w)
		next(rw, req)
		statusCode := rw.StatusCode

		// Tracing
		// https://github.com/opentracing/specification/blob/master/semantic_conventions.md
		span.SetTag("http.proto", proto)
		ext.HTTPMethod.Set(span, method)
		ext.HTTPUrl.Set(span, url)
		ext.HTTPStatusCode.Set(span, uint16(statusCode))
		/* span.LogFields(
			opentracingLog.String("key", value),
		) */
	}
}
