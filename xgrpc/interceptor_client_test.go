package xgrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/moorara/observe/log"
	"github.com/moorara/observe/metrics"
	"github.com/moorara/observe/request"
	"github.com/moorara/observe/trace"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/prometheus/client_golang/prometheus"
	promModel "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func extractSpanContext(ctx context.Context, tracer opentracing.Tracer) opentracing.SpanContext {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		carrier := &metadataTextMap{md}
		parentSpanContext, err := tracer.Extract(opentracing.TextMap, carrier)
		if err == nil {
			return parentSpanContext
		}
	}

	return nil
}

func extractRequestMetadata(ctx context.Context) (string, string) {
	var id, name string

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		if vals := md.Get(requestIDKey); len(vals) > 0 {
			id = vals[0]
		}

		if vals := md.Get(clientNameKey); len(vals) > 0 {
			name = vals[0]
		}
	}

	return id, name
}

func TestClientInterceptorOptions(t *testing.T) {
	logger := log.NewVoidLogger()
	// mf := metrics.NewFactory(metrics.FactoryOptions{Registerer: prometheus.NewRegistry()})
	tracer := mocktracer.New()

	tests := []struct {
		name                      string
		clientInterceptor         ClientInterceptor
		opt                       ClientInterceptorOption
		expectedClientInterceptor ClientInterceptor
	}{
		{
			"ClientLogging",
			ClientInterceptor{},
			ClientLogging(logger),
			ClientInterceptor{
				logger: logger,
			},
		},
		/* {
			"ClientMetrics",
			ClientInterceptor{},
			ClientMetrics(mf),
			ClientInterceptor{},
		}, */
		{
			"ClientTracing",
			ClientInterceptor{},
			ClientTracing(tracer),
			ClientInterceptor{
				tracer: tracer,
			},
		},
		{
			"ClientFilter",
			ClientInterceptor{},
			ClientFilter("testPB", "Manager", "Ping"),
			ClientInterceptor{
				filters: []filter{
					{"testPB", "Manager", "Ping"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.opt(&tc.clientInterceptor)

			assert.Equal(t, tc.expectedClientInterceptor, tc.clientInterceptor)
		})
	}
}

func TestNewClientInterceptor(t *testing.T) {
	logger := log.NewLogger(log.Options{
		Level:       "info",
		Name:        "logger",
		Environment: "test",
	})

	promReg := prometheus.NewRegistry()
	mFac := metrics.NewFactory(metrics.FactoryOptions{Registerer: promReg})

	tracer, closer, _ := trace.NewTracer(trace.Options{})
	defer closer.Close()

	tests := []struct {
		name       string
		clientName string
		opts       []ClientInterceptorOption
	}{
		{
			"Default",
			"test-driver",
			[]ClientInterceptorOption{
				ClientLogging(logger),
				ClientMetrics(mFac),
				ClientTracing(tracer),
			},
		},
		{
			"WithMocks",
			"test-driver",
			[]ClientInterceptorOption{
				ClientLogging(log.NewVoidLogger()),
				ClientMetrics(metrics.NewFactory(metrics.FactoryOptions{})),
				ClientTracing(mocktracer.New()),
			},
		},
		{
			"WithFilters",
			"test-driver",
			[]ClientInterceptorOption{
				ClientFilter("teamPB", "Manager", "Ping"),
				ClientFilter("groupPB", "Manager", "Ping"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ci := NewClientInterceptor(tc.clientName, tc.opts...)

			assert.NotNil(t, ci)
			assert.Equal(t, tc.clientName, ci.name)
		})
	}
}

func TestInjectSpan(t *testing.T) {
	tracer := mocktracer.New()
	ctx := context.Background()
	ctxWithMD := metadata.NewOutgoingContext(ctx, metadata.Pairs("key", "value"))

	tests := []struct {
		name     string
		tracer   opentracing.Tracer
		ctx      context.Context
		span     opentracing.Span
		expected bool
	}{
		{
			name:     "WithoutMetadata",
			tracer:   tracer,
			ctx:      ctx,
			span:     tracer.StartSpan("test-span"),
			expected: true,
		},
		{
			name:     "WithMetadata",
			tracer:   tracer,
			ctx:      ctxWithMD,
			span:     tracer.StartSpan("test-span"),
			expected: true,
		},
		{
			name:     "InjectFails",
			tracer:   tracer,
			ctx:      ctx,
			span:     &mockSpan{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := &ClientInterceptor{
				tracer: tc.tracer,
			}

			ctx := i.injectSpan(tc.ctx, tc.span)

			injectedSpanContext := extractSpanContext(ctx, tc.tracer)
			assert.Equal(t, tc.expected, injectedSpanContext != nil)
		})
	}
}

func TestInjectRequestMetadata(t *testing.T) {
	ctx := context.Background()
	ctxWithMD := metadata.NewOutgoingContext(ctx, metadata.Pairs("key", "value"))

	tests := []struct {
		name       string
		ctx        context.Context
		requestID  string
		clientName string
	}{
		{
			name:       "WithoutMetadata",
			ctx:        ctx,
			requestID:  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			clientName: "test-driver",
		},
		{
			name:       "WithMetadata",
			ctx:        ctxWithMD,
			requestID:  "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			clientName: "test-driver",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			i := &ClientInterceptor{}

			ctx := i.injectRequestMetadata(tc.ctx, tc.requestID, tc.clientName)

			requestID, clientName := extractRequestMetadata(ctx)
			assert.Equal(t, tc.requestID, requestID)
			assert.Equal(t, tc.clientName, clientName)
		})
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	tests := []struct {
		name            string
		clientName      string
		filters         []filter
		parentSpan      opentracing.Span
		requestID       string
		ctx             context.Context
		method          string
		req             interface{}
		res             interface{}
		cc              *grpc.ClientConn
		opts            []grpc.CallOption
		mockDelay       time.Duration
		mockRespError   error
		verify          bool
		expectedPackage string
		expectedService string
		expectedMethod  string
		expectedStream  string
		expectedSuccess bool
	}{
		{
			name:            "InvalidMethod",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			method:          "",
			req:             nil,
			res:             nil,
			cc:              nil,
			opts:            nil,
			mockDelay:       0,
			mockRespError:   nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "PackageFilterMatches",
			clientName:      "service-name",
			filters:         []filter{{"package", "", ""}},
			parentSpan:      nil,
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       0,
			mockRespError:   nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "ServiceFilterMatches",
			clientName:      "service-name",
			filters:         []filter{{"package", "service", ""}},
			parentSpan:      nil,
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       0,
			mockRespError:   nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "MethodFilterMatches",
			clientName:      "service-name",
			filters:         []filter{{"package", "service", "method"}},
			parentSpan:      nil,
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       0,
			mockRespError:   nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "InvokerSucceeds",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "false",
			expectedSuccess: true,
		},
		{
			name:            "InvokerFails",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   errors.New("error on grpc method"),
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "false",
			expectedSuccess: false,
		},
		{
			name:            "InvokerSucceeds",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "false",
			expectedSuccess: true,
		},
		{
			name:            "InvokerSucceedsWithMetadata",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             metadata.NewOutgoingContext(context.Background(), metadata.Pairs("key", "value")),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "false",
			expectedSuccess: true,
		},
		{
			name:            "InvokerSucceedsWithParentSpan",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      mocktracer.New().StartSpan("parent-span"),
			requestID:       "",
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "false",
			expectedSuccess: true,
		},
		{
			name:            "InvokerSucceedsWithRequestID",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			ctx:             context.Background(),
			method:          "/package.service/method",
			req:             nil,
			res:             nil,
			cc:              &grpc.ClientConn{},
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "false",
			expectedSuccess: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buff := &bytes.Buffer{}
			var injectedSpanContext opentracing.SpanContext
			var injectedRequestID, injectedClientName string

			logger := log.NewLogger(log.Options{Writer: buff})
			promReg := prometheus.NewRegistry()
			mf := metrics.NewFactory(metrics.FactoryOptions{Registerer: promReg})
			tracer := mocktracer.New()

			opts := []ClientInterceptorOption{
				ClientLogging(logger),
				ClientMetrics(mf),
				ClientTracing(tracer),
			}

			// Apply filters if any
			for _, f := range tc.filters {
				opts = append(opts, ClientFilter(f.pkg, f.service, f.method))
			}

			// Create the interceptor
			i := NewClientInterceptor(tc.clientName, opts...)
			assert.NotNil(t, i)

			if tc.parentSpan != nil {
				tc.ctx = opentracing.ContextWithSpan(tc.ctx, tc.parentSpan)
			}

			if tc.requestID != "" {
				tc.ctx = request.ContextWithID(tc.ctx, tc.requestID)
			}

			invoker := func(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				time.Sleep(tc.mockDelay)
				injectedSpanContext = extractSpanContext(ctx, tracer)
				injectedRequestID, injectedClientName = extractRequestMetadata(ctx)
				return tc.mockRespError
			}

			err := i.UnaryInterceptor(tc.ctx, tc.method, tc.req, tc.res, tc.cc, invoker, tc.opts...)
			assert.Equal(t, tc.mockRespError, err)

			if tc.verify {
				// Verify request id
				if tc.requestID != "" {
					assert.Equal(t, tc.requestID, injectedRequestID)
				} else {
					assert.NotEmpty(t, injectedRequestID)
				}

				// Verify client name
				assert.Equal(t, tc.clientName, injectedClientName)

				// Verify logs

				var log map[string]interface{}
				err := json.NewDecoder(buff).Decode(&log)
				assert.NoError(t, err)
				assert.Equal(t, clientKind, log["grpc.kind"])
				assert.Equal(t, tc.expectedPackage, log["grpc.package"])
				assert.Equal(t, tc.expectedService, log["grpc.service"])
				assert.Equal(t, tc.expectedMethod, log["grpc.method"])
				assert.Equal(t, tc.expectedStream, log["grpc.stream"])
				assert.Equal(t, tc.expectedSuccess, log["grpc.success"])
				assert.NotEmpty(t, log["responseTime"])
				assert.NotEmpty(t, log["message"])

				if tc.mockRespError != nil {
					assert.Equal(t, tc.mockRespError.Error(), log["grpc.error"])
				}

				if tc.requestID != "" {
					assert.Equal(t, tc.requestID, log["requestId"])
				} else {
					assert.NotEmpty(t, log["requestId"])
				}

				// Verify metrics

				verifyLabels := func(labels []*promModel.LabelPair) {
					for _, l := range labels {
						switch *l.Name {
						case "package":
							assert.Equal(t, tc.expectedPackage, *l.Value)
						case "service":
							assert.Equal(t, tc.expectedService, *l.Value)
						case "method":
							assert.Equal(t, tc.expectedMethod, *l.Value)
						case "stream":
							assert.Equal(t, tc.expectedStream, *l.Value)
						case "success":
							assert.Equal(t, strconv.FormatBool(tc.expectedSuccess), *l.Value)
						}
					}
				}

				metricFamilies, err := promReg.Gather()
				assert.NoError(t, err)

				for _, metricFamily := range metricFamilies {
					switch *metricFamily.Name {
					case clientGaugeMetricName:
						assert.Equal(t, promModel.MetricType_GAUGE, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					case clientCounterMetricName:
						assert.Equal(t, promModel.MetricType_COUNTER, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					case clientHistogramMetricName:
						assert.Equal(t, promModel.MetricType_HISTOGRAM, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					case clientSummaryMetricName:
						assert.Equal(t, promModel.MetricType_SUMMARY, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					}
				}

				// Verify traces

				span := tracer.FinishedSpans()[0]
				assert.Equal(t, injectedSpanContext, span.Context())
				assert.Equal(t, clientSpanName, span.OperationName)
				assert.Equal(t, ext.SpanKindEnum("client"), span.Tag("span.kind"))
				assert.Equal(t, tc.expectedPackage, span.Tag("grpc.package"))
				assert.Equal(t, tc.expectedService, span.Tag("grpc.service"))
				assert.Equal(t, tc.expectedMethod, span.Tag("grpc.method"))
				assert.Equal(t, tc.expectedStream, span.Tag("grpc.stream"))
				assert.Equal(t, tc.expectedSuccess, span.Tag("grpc.success"))

				if tc.parentSpan != nil {
					parentSpan, ok := tc.parentSpan.(*mocktracer.MockSpan)
					assert.True(t, ok)
					assert.Equal(t, parentSpan.SpanContext.SpanID, span.ParentID)
					assert.Equal(t, parentSpan.SpanContext.TraceID, span.SpanContext.TraceID)
				}

				spanLogs := span.Logs()
				if tc.mockRespError != nil {
					lf := spanLogs[0].Fields[0]
					assert.True(t, span.Tag("error").(bool))
					assert.Equal(t, "grpc.error", lf.Key)
					assert.Equal(t, tc.mockRespError.Error(), lf.ValueString)
				}
			}
		})
	}
}

func TestStreamClientInterceptor(t *testing.T) {
	tests := []struct {
		name            string
		clientName      string
		filters         []filter
		parentSpan      opentracing.Span
		requestID       string
		ctx             context.Context
		desc            *grpc.StreamDesc
		cc              *grpc.ClientConn
		method          string
		opts            []grpc.CallOption
		mockDelay       time.Duration
		mockRespError   error
		mockRespCS      grpc.ClientStream
		verify          bool
		expectedPackage string
		expectedService string
		expectedMethod  string
		expectedStream  string
		expectedSuccess bool
	}{
		{
			name:            "InvalidMethod",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			desc:            nil,
			cc:              nil,
			method:          "",
			opts:            nil,
			mockDelay:       0,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "PackageFilterMatches",
			clientName:      "service-name",
			filters:         []filter{{"package", "", ""}},
			parentSpan:      nil,
			ctx:             context.Background(),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       0,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "ServiceFilterMatches",
			clientName:      "service-name",
			filters:         []filter{{"package", "service", ""}},
			parentSpan:      nil,
			ctx:             context.Background(),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       0,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "MethodFilterMatches",
			clientName:      "service-name",
			filters:         []filter{{"package", "service", "method"}},
			parentSpan:      nil,
			ctx:             context.Background(),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       0,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          false,
			expectedPackage: "",
			expectedService: "",
			expectedMethod:  "",
			expectedStream:  "",
			expectedSuccess: false,
		},
		{
			name:            "StreamerFails",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   errors.New("error on grpc method"),
			mockRespCS:      nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "true",
			expectedSuccess: false,
		},
		{
			name:            "StreamerSucceeds",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "true",
			expectedSuccess: true,
		},
		{
			name:            "StreamerSucceedsWithMetadata",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             metadata.NewOutgoingContext(context.Background(), metadata.Pairs("key", "value")),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "true",
			expectedSuccess: true,
		},
		{
			name:            "StreamerSucceedsWithParentSpan",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      mocktracer.New().StartSpan("parent-span"),
			requestID:       "",
			ctx:             context.Background(),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "true",
			expectedSuccess: true,
		},
		{
			name:            "StreamerSucceedsWithRequestID",
			clientName:      "service-name",
			filters:         nil,
			parentSpan:      nil,
			requestID:       "",
			ctx:             context.Background(),
			desc:            &grpc.StreamDesc{},
			cc:              &grpc.ClientConn{},
			method:          "/package.service/method",
			opts:            []grpc.CallOption{},
			mockDelay:       10 * time.Millisecond,
			mockRespError:   nil,
			mockRespCS:      nil,
			verify:          true,
			expectedPackage: "package",
			expectedService: "service",
			expectedMethod:  "method",
			expectedStream:  "true",
			expectedSuccess: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buff := &bytes.Buffer{}
			var injectedSpanContext opentracing.SpanContext
			var injectedRequestID, injectedClientName string

			logger := log.NewLogger(log.Options{Writer: buff})
			promReg := prometheus.NewRegistry()
			mf := metrics.NewFactory(metrics.FactoryOptions{Registerer: promReg})
			tracer := mocktracer.New()

			opts := []ClientInterceptorOption{
				ClientLogging(logger),
				ClientMetrics(mf),
				ClientTracing(tracer),
			}

			// Apply filters if any
			for _, f := range tc.filters {
				opts = append(opts, ClientFilter(f.pkg, f.service, f.method))
			}

			// Create the interceptor
			i := NewClientInterceptor(tc.clientName, opts...)
			assert.NotNil(t, i)

			if tc.parentSpan != nil {
				tc.ctx = opentracing.ContextWithSpan(tc.ctx, tc.parentSpan)
			}

			if tc.requestID != "" {
				tc.ctx = request.ContextWithID(tc.ctx, tc.requestID)
			}

			streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				time.Sleep(tc.mockDelay)
				injectedSpanContext = extractSpanContext(ctx, tracer)
				injectedRequestID, injectedClientName = extractRequestMetadata(ctx)
				return tc.mockRespCS, tc.mockRespError
			}

			cs, err := i.StreamInterceptor(tc.ctx, tc.desc, tc.cc, tc.method, streamer, tc.opts...)
			assert.Equal(t, tc.mockRespError, err)
			assert.Equal(t, tc.mockRespCS, cs)

			if tc.verify {
				// Verify request id
				if tc.requestID != "" {
					assert.Equal(t, tc.requestID, injectedRequestID)
				} else {
					assert.NotEmpty(t, injectedRequestID)
				}

				// Verify client name
				assert.Equal(t, tc.clientName, injectedClientName)

				// Verify logs

				var log map[string]interface{}
				err := json.NewDecoder(buff).Decode(&log)
				assert.NoError(t, err)
				assert.Equal(t, clientKind, log["grpc.kind"])
				assert.Equal(t, tc.expectedPackage, log["grpc.package"])
				assert.Equal(t, tc.expectedService, log["grpc.service"])
				assert.Equal(t, tc.expectedMethod, log["grpc.method"])
				assert.Equal(t, tc.expectedStream, log["grpc.stream"])
				assert.Equal(t, tc.expectedSuccess, log["grpc.success"])
				assert.NotEmpty(t, log["responseTime"])
				assert.NotEmpty(t, log["message"])

				if tc.mockRespError != nil {
					assert.Equal(t, tc.mockRespError.Error(), log["grpc.error"])
				}

				if tc.requestID != "" {
					assert.Equal(t, tc.requestID, log["requestId"])
				} else {
					assert.NotEmpty(t, log["requestId"])
				}

				// Verify metrics

				verifyLabels := func(labels []*promModel.LabelPair) {
					for _, l := range labels {
						switch *l.Name {
						case "package":
							assert.Equal(t, tc.expectedPackage, *l.Value)
						case "service":
							assert.Equal(t, tc.expectedService, *l.Value)
						case "method":
							assert.Equal(t, tc.expectedMethod, *l.Value)
						case "stream":
							assert.Equal(t, tc.expectedStream, *l.Value)
						case "success":
							assert.Equal(t, strconv.FormatBool(tc.expectedSuccess), *l.Value)
						}
					}
				}

				metricFamilies, err := promReg.Gather()
				assert.NoError(t, err)

				for _, metricFamily := range metricFamilies {
					switch *metricFamily.Name {
					case clientGaugeMetricName:
						assert.Equal(t, promModel.MetricType_GAUGE, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					case clientCounterMetricName:
						assert.Equal(t, promModel.MetricType_COUNTER, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					case clientHistogramMetricName:
						assert.Equal(t, promModel.MetricType_HISTOGRAM, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					case clientSummaryMetricName:
						assert.Equal(t, promModel.MetricType_SUMMARY, *metricFamily.Type)
						verifyLabels(metricFamily.Metric[0].Label)
					}
				}

				// Verify traces

				span := tracer.FinishedSpans()[0]
				assert.Equal(t, injectedSpanContext, span.Context())
				assert.Equal(t, clientSpanName, span.OperationName)
				assert.Equal(t, ext.SpanKindEnum("client"), span.Tag("span.kind"))
				assert.Equal(t, tc.expectedPackage, span.Tag("grpc.package"))
				assert.Equal(t, tc.expectedService, span.Tag("grpc.service"))
				assert.Equal(t, tc.expectedMethod, span.Tag("grpc.method"))
				assert.Equal(t, tc.expectedStream, span.Tag("grpc.stream"))
				assert.Equal(t, tc.expectedSuccess, span.Tag("grpc.success"))

				if tc.parentSpan != nil {
					parentSpan, ok := tc.parentSpan.(*mocktracer.MockSpan)
					assert.True(t, ok)
					assert.Equal(t, parentSpan.SpanContext.SpanID, span.ParentID)
					assert.Equal(t, parentSpan.SpanContext.TraceID, span.SpanContext.TraceID)
				}

				spanLogs := span.Logs()
				if tc.mockRespError != nil {
					lf := spanLogs[0].Fields[0]
					assert.True(t, span.Tag("error").(bool))
					assert.Equal(t, "grpc.error", lf.Key)
					assert.Equal(t, tc.mockRespError.Error(), lf.ValueString)
				}
			}
		})
	}
}
