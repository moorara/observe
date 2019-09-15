package request

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	requestID := NewID()

	u, err := uuid.Parse(requestID)
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
}

func TestContextWithID(t *testing.T) {
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
			ctx := ContextWithID(tc.ctx, tc.requestID)

			assert.Equal(t, tc.requestID, ctx.Value(requestIDContextKey))
		})
	}
}

func TestIDFromContext(t *testing.T) {
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
			context.WithValue(context.Background(), requestIDContextKey, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			true,
			"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			requestID, ok := IDFromContext(tc.ctx)

			assert.Equal(t, tc.expectedOK, ok)
			assert.Equal(t, tc.expectedRequestID, requestID)
		})
	}
}
