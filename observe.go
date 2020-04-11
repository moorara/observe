// Package observe is used for implementing three pillars of observability using OpenTelemetry API.
// It provides structured logging, metrics, and tracing in one package.
//
// Logging:
// You can either create an instance logger or use the singleton (global) logger.
// observe.Logger is concurrently safe to be used by multiple goroutines.
//
// Metrics:
//
// Tracing:
//
package observe
