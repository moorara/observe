package log

import (
	"context"
	"io"
	"os"
	"strings"

	kitLog "github.com/go-kit/kit/log"
	kitLevel "github.com/go-kit/kit/log/level"
)

const (
	withCallerDepth      = 6
	instanceCallerDepth  = 7
	singletonCallerDepth = 8
)

// Format is the type for output format.
type Format int

const (
	// JSON represents a json logger
	JSON Format = iota
	// Logfmt represents logfmt logger
	Logfmt
)

// Level is the type for logging level.
type Level int

const (
	// NoneLevel log
	NoneLevel Level = iota
	// ErrorLevel log
	ErrorLevel
	// WarnLevel log
	WarnLevel
	// InfoLevel log
	InfoLevel
	// DebugLevel log
	DebugLevel
)

func stringToLevel(level string) Level {
	switch strings.ToLower(level) {
	case "none":
		return NoneLevel
	case "error":
		return ErrorLevel
	case "warn":
		return WarnLevel
	case "info":
		return InfoLevel
	case "debug":
		return DebugLevel
	default:
		return InfoLevel
	}
}

// Options contains optional options for Logger.
type Options struct {
	callerDepth int
	Name        string
	Environment string
	Region      string
	Level       string
	Format      Format
	Writer      io.Writer
}

func createKitLogger(opts Options) kitLog.Logger {
	var kitLogger kitLog.Logger

	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch opts.Format {
	case Logfmt:
		kitLogger = kitLog.NewLogfmtLogger(opts.Writer)
	case JSON:
		fallthrough
	default:
		kitLogger = kitLog.NewJSONLogger(opts.Writer)
	}

	// This is not required since SwapLogger uses a SyncLogger and can be used concurrently
	// kitLogger = kitLog.NewSyncLogger(kitLogger)

	if opts.callerDepth == 0 {
		opts.callerDepth = instanceCallerDepth
	}

	kitLogger = kitLog.With(kitLogger,
		"caller", kitLog.Caller(opts.callerDepth),
		"timestamp", kitLog.DefaultTimestampUTC,
	)

	if opts.Name != "" {
		kitLogger = kitLog.With(kitLogger, "logger", opts.Name)
	}

	if opts.Environment != "" {
		kitLogger = kitLog.With(kitLogger, "environment", opts.Environment)
	}

	if opts.Region != "" {
		kitLogger = kitLog.With(kitLogger, "region", opts.Region)
	}

	switch strings.ToLower(opts.Level) {
	case "none":
		kitLogger = kitLevel.NewFilter(kitLogger, kitLevel.AllowNone())
	case "error":
		kitLogger = kitLevel.NewFilter(kitLogger, kitLevel.AllowError())
	case "warn":
		kitLogger = kitLevel.NewFilter(kitLogger, kitLevel.AllowWarn())
	case "info":
		kitLogger = kitLevel.NewFilter(kitLogger, kitLevel.AllowInfo())
	case "debug":
		kitLogger = kitLevel.NewFilter(kitLogger, kitLevel.AllowDebug())
	default:
		kitLogger = kitLevel.NewFilter(kitLogger, kitLevel.AllowInfo())
	}

	return kitLogger
}

// Logger wraps a go-kit Logger.
type Logger struct {
	Level  Level
	Logger *kitLog.SwapLogger
}

// NewLogger creates a new logger.
func NewLogger(opts Options) *Logger {
	logger := &Logger{
		Level:  stringToLevel(opts.Level),
		Logger: new(kitLog.SwapLogger),
	}

	kitLogger := createKitLogger(opts)
	logger.Logger.Swap(kitLogger)

	return logger
}

// NewVoidLogger creates a void logger for testing purposes.
func NewVoidLogger() *Logger {
	logger := &Logger{
		Logger: new(kitLog.SwapLogger),
	}

	kitLogger := kitLog.NewNopLogger()
	logger.Logger.Swap(kitLogger)

	return logger
}

// With returns a new logger that always logs a set of key-value pairs (context).
func (l *Logger) With(kv ...interface{}) *Logger {
	logger := &Logger{
		Level:  l.Level,
		Logger: new(kitLog.SwapLogger),
	}

	kitLogger := kitLog.With(l.Logger, kv...)
	kitLogger = kitLog.With(kitLogger, "caller", kitLog.Caller(withCallerDepth))
	logger.Logger.Swap(kitLogger)

	return logger
}

// SetOptions resets a logger with new options.
func (l *Logger) SetOptions(opts Options) {
	kitLogger := createKitLogger(opts)
	l.Logger.Swap(kitLogger)
}

// Debug logs in debug level.
func (l *Logger) Debug(kv ...interface{}) error {
	return kitLevel.Debug(l.Logger).Log(kv...)
}

// Info logs in debug level.
func (l *Logger) Info(kv ...interface{}) error {
	return kitLevel.Info(l.Logger).Log(kv...)
}

// Warn logs in debug level.
func (l *Logger) Warn(kv ...interface{}) error {
	return kitLevel.Warn(l.Logger).Log(kv...)
}

// Error logs in debug level.
func (l *Logger) Error(kv ...interface{}) error {
	return kitLevel.Error(l.Logger).Log(kv...)
}

// The singleton logger.
var singleton = NewLogger(Options{
	Name:        "singleton",
	callerDepth: singletonCallerDepth,
})

// SetOptions set optional options for singleton logger.
func SetOptions(opts Options) {
	opts.callerDepth = 8
	singleton.SetOptions(opts)
	singleton.Level = stringToLevel(opts.Level)
}

// Debug logs a debug-level event using singleton logger.
func Debug(kv ...interface{}) error {
	return singleton.Debug(kv...)
}

// Info logs an info-level event using singleton logger.
func Info(kv ...interface{}) error {
	return singleton.Info(kv...)
}

// Warn logs a warn-level event using singleton logger.
func Warn(kv ...interface{}) error {
	return singleton.Warn(kv...)
}

// Error logs an error-level event using singleton logger.
func Error(kv ...interface{}) error {
	return singleton.Error(kv...)
}

// contextKey is the type for the keys added to context.
type contextKey string

const loggerContextKey = contextKey("logger")

// ContextWithLogger returns a new context that holds a reference to the logger.
func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

// LoggerFromContext returns a logger set on a context.
// If no logger found on the context, the singleton logger will be returned.
func LoggerFromContext(ctx context.Context) *Logger {
	val := ctx.Value(loggerContextKey)
	if logger, ok := val.(*Logger); ok {
		return logger
	}

	return singleton
}
