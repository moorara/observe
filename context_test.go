package observe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestContextWithUUID(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		requestID string
	}{
		{
			"OK",
			context.Background(),
			"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ContextWithUUID(tc.ctx, tc.requestID)

			assert.Equal(t, tc.requestID, ctx.Value(uuidContextKey))
		})
	}
}

func TestUUIDFromContext(t *testing.T) {
	tests := []struct {
		name              string
		ctx               context.Context
		expectedOK        bool
		expectedRequestID string
	}{
		{
			"NoRequestID",
			context.Background(),
			false,
			"",
		},
		{
			"WithRequestID",
			context.WithValue(context.Background(), uuidContextKey, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			true,
			"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestID, ok := UUIDFromContext(tc.ctx)

			assert.Equal(t, tc.expectedOK, ok)
			assert.Equal(t, tc.expectedRequestID, requestID)
		})
	}
}

func TestContextWithLogger(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		logger *zap.Logger
	}{
		{
			name:   "NoLogger",
			ctx:    context.Background(),
			logger: nil,
		},
		{
			name:   "WithLogger",
			ctx:    context.Background(),
			logger: zap.NewExample(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ContextWithLogger(tc.ctx, tc.logger)
			logger := ctx.Value(loggerContextKey)

			assert.Equal(t, tc.logger, logger)
		})
	}
}

func TestLoggerFromContext(t *testing.T) {
	logger := zap.NewExample()

	tests := []struct {
		name           string
		ctx            context.Context
		expectedLogger *zap.Logger
	}{
		{
			name:           "NoLogger",
			ctx:            context.Background(),
			expectedLogger: Logger,
		},
		{
			name:           "WithLogger",
			ctx:            context.WithValue(context.Background(), loggerContextKey, logger),
			expectedLogger: logger,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := LoggerFromContext(tc.ctx)

			assert.Equal(t, tc.expectedLogger, logger)
		})
	}
}
