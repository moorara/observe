package request

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is the type for the keys added to context.
type contextKey string

const requestIDContextKey = contextKey("RequestID")

// NewID creates a new request id.
func NewID() string {
	return uuid.New().String()
}

// ContextWithID creates a new context with a request id.
func ContextWithID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

// IDFromContext retrieves a request id from a context.
func IDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDContextKey).(string)
	return requestID, ok
}
