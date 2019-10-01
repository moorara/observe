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
	opentracingLog "github.com/opentracing/opentracing-go/log"
)

const (
	clientKind                = "client"
	clientSpanName            = "http-client-request"
	clientGaugeMetricName     = "http_client_requests"
	clientCounterMetricName   = "http_client_requests_total"
	clientHistogramMetricName = "http_client_request_duration_seconds"
	clientSummaryMetricName   = "http_client_request_duration_quantiles_seconds"
)

// Doer is the interface for standard http.Client Do method.
type Doer func(*http.Request) (*http.Response, error)

// ClientMiddleware is an http client middleware for logging, metrics, tracing, etc.
type ClientMiddleware struct {
	logger  *log.Logger
	metrics *metrics.RequestMetrics
	tracer  opentracing.Tracer
}

// ClientMiddlewareOption sets optional parameters for client middleware.
type ClientMiddlewareOption func(*ClientMiddleware)

// ClientLogging is the option for client middleware to enable logging for every request.
func ClientLogging(logger *log.Logger) ClientMiddlewareOption {
	return func(i *ClientMiddleware) {
		i.logger = logger
	}
}

// ClientMetrics is the option for client middleware to enable metrics for every request.
func ClientMetrics(mf *metrics.Factory) ClientMiddlewareOption {
	metrics := &metrics.RequestMetrics{
		ReqGauge:        mf.Gauge(clientGaugeMetricName, "gauge metric for number of active client-side http requests", []string{"method", "url"}),
		ReqCounter:      mf.Counter(clientCounterMetricName, "counter metric for total number of client-side http requests", []string{"method", "url", "statusCode", "statusClass"}),
		ReqDurationHist: mf.Histogram(clientHistogramMetricName, "histogram metric for duration of client-side http requests in seconds", []string{"method", "url", "statusCode", "statusClass"}),
		ReqDurationSumm: mf.Summary(clientSummaryMetricName, "summary metric for duration of client-side http requests in seconds", []string{"method", "url", "statusCode", "statusClass"}),
	}

	return func(i *ClientMiddleware) {
		i.metrics = metrics
	}
}

// ClientTracing is the option for client middleware to enable tracing for every request.
func ClientTracing(tracer opentracing.Tracer) ClientMiddlewareOption {
	return func(i *ClientMiddleware) {
		i.tracer = tracer
	}
}

// NewClientMiddleware creates a new instance of http client middleware.
func NewClientMiddleware(opts ...ClientMiddlewareOption) *ClientMiddleware {
	cm := &ClientMiddleware{}
	for _, opt := range opts {
		opt(cm)
	}

	return cm
}

// RequestID ensures outgoing requests have unique ids.
// This middleware ensures the request headers and context have a unique id.
// A new request id will be generated if needed.
func (m *ClientMiddleware) RequestID(next Doer) Doer {
	return func(r *http.Request) (*http.Response, error) {
		// Ensure the request context has a request id
		requestID, ok := request.IDFromContext(r.Context())
		if !ok || requestID == "" {
			requestID = r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = request.NewID()
			}
			r = r.WithContext(request.ContextWithID(r.Context(), requestID))
		}

		// Ensure request headers have the request id
		// Using Set will override any previously added id
		// Using Add ensures an id added to headers will have a higher priority over an id added to context
		r.Header.Add(requestIDHeader, requestID)

		// Call the next request doer
		return next(r)
	}
}

// Logging takes care of logging for outgoing http requests.
// Request id will be read from reqeust context if present.
func (m *ClientMiddleware) Logging(next Doer) Doer {
	return func(r *http.Request) (*http.Response, error) {
		proto := r.Proto
		method := r.Method
		url := r.URL.Path

		// Get request id from context
		requestID, _ := request.IDFromContext(r.Context())

		// Call the next request doer
		start := time.Now()
		res, err := next(r)
		duration := time.Since(start).Seconds()

		var statusCode int
		var statusClass string

		if err != nil {
			statusCode = -1
			statusClass = ""
		} else {
			statusCode = res.StatusCode
			statusClass = fmt.Sprintf("%dxx", statusCode/100)
		}

		pairs := []interface{}{
			"http.kind", clientKind,
			"req.proto", proto,
			"req.method", method,
			"req.url", url,
			"res.statusCode", statusCode,
			"res.statusClass", statusClass,
			"responseTime", duration,
			"message", fmt.Sprintf("%s %s %d %f", method, url, statusCode, duration),
		}

		if requestID != "" {
			pairs = append(pairs, "requestId", requestID)
		}

		// Logging
		switch {
		case statusCode >= 500:
			m.logger.ErrorKV(pairs...)
		case statusCode >= 400:
			m.logger.WarnKV(pairs...)
		case statusCode >= 100:
			fallthrough
		default:
			m.logger.InfoKV(pairs...)
		}

		return res, err
	}
}

// Metrics takes care of metrics for outgoing http requests.
func (m *ClientMiddleware) Metrics(next Doer) Doer {
	return func(r *http.Request) (*http.Response, error) {
		method := r.Method
		url := r.URL.Path

		// Increment guage metric
		m.metrics.ReqGauge.WithLabelValues(method, url).Inc()

		// Call the next request doer
		start := time.Now()
		res, err := next(r)
		duration := time.Since(start).Seconds()

		var statusCode int
		var statusClass string

		if err != nil {
			statusCode = -1
			statusClass = ""
		} else {
			statusCode = res.StatusCode
			statusClass = fmt.Sprintf("%dxx", statusCode/100)
		}

		// Metrics
		statusText := strconv.Itoa(statusCode)
		m.metrics.ReqGauge.WithLabelValues(method, url).Dec()
		m.metrics.ReqCounter.WithLabelValues(method, url, statusText, statusClass).Inc()
		m.metrics.ReqDurationHist.WithLabelValues(method, url, statusText, statusClass).Observe(duration)
		m.metrics.ReqDurationSumm.WithLabelValues(method, url, statusText, statusClass).Observe(duration)

		return res, err
	}
}

func (m *ClientMiddleware) createSpan(ctx context.Context) opentracing.Span {
	var span opentracing.Span

	// Get trace information from the context if passed
	parentSpan := opentracing.SpanFromContext(ctx)

	if parentSpan == nil {
		span = m.tracer.StartSpan(clientSpanName)
	} else {
		span = m.tracer.StartSpan(clientSpanName, opentracing.ChildOf(parentSpan.Context()))
	}

	return span
}

func (m *ClientMiddleware) injectSpan(req *http.Request, span opentracing.Span) {
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	err := m.tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	if err != nil {
		span.LogFields(
			opentracingLog.Error(err),
			opentracingLog.String("message", "Tracer.Inject() failed"),
		)
	}
}

// Tracing takes care of tracing for outgoing http requests.
// Trace information will be read from reqeust context if present.
func (m *ClientMiddleware) Tracing(next Doer) Doer {
	return func(r *http.Request) (*http.Response, error) {
		proto := r.Proto
		method := r.Method
		url := r.URL.Path

		// Create a new span and propagate the current trace
		span := m.createSpan(r.Context())
		defer span.Finish()
		m.injectSpan(r, span)

		// Call the next request doer
		res, err := next(r)

		var statusCode int
		if err != nil {
			statusCode = -1
		} else {
			statusCode = res.StatusCode
		}

		// Tracing
		// https://github.com/opentracing/specification/blob/master/semantic_conventions.md
		span.SetTag("http.proto", proto)
		ext.HTTPMethod.Set(span, method)
		ext.HTTPUrl.Set(span, url)
		ext.HTTPStatusCode.Set(span, uint16(statusCode))
		/* span.LogFields(
			opentracingLog.String("key", value),
		) */

		return res, err
	}
}
