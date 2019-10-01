package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	kitLog "github.com/go-kit/kit/log"
	kitLevel "github.com/go-kit/kit/log/level"
)

const (
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

func createBaseLogger(opts Options) kitLog.Logger {
	var base kitLog.Logger

	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch opts.Format {
	case Logfmt:
		base = kitLog.NewLogfmtLogger(opts.Writer)
	case JSON:
		fallthrough
	default:
		base = kitLog.NewJSONLogger(opts.Writer)
	}

	// This is not required since SwapLogger uses a SyncLogger and can be used concurrently
	// base = kitLog.NewSyncLogger(base)

	if opts.callerDepth == 0 {
		opts.callerDepth = instanceCallerDepth
	}

	base = kitLog.With(base,
		"caller", kitLog.Caller(opts.callerDepth),
		"timestamp", kitLog.DefaultTimestampUTC,
	)

	if opts.Name != "" {
		base = kitLog.With(base, "logger", opts.Name)
	}

	if opts.Environment != "" {
		base = kitLog.With(base, "environment", opts.Environment)
	}

	if opts.Region != "" {
		base = kitLog.With(base, "region", opts.Region)
	}

	return base
}

func createFilteredLogger(base kitLog.Logger, level Level) kitLog.Logger {
	var filtered kitLog.Logger

	switch level {
	case NoneLevel:
		filtered = kitLevel.NewFilter(base, kitLevel.AllowNone())
	case ErrorLevel:
		filtered = kitLevel.NewFilter(base, kitLevel.AllowError())
	case WarnLevel:
		filtered = kitLevel.NewFilter(base, kitLevel.AllowWarn())
	case InfoLevel:
		filtered = kitLevel.NewFilter(base, kitLevel.AllowInfo())
	case DebugLevel:
		filtered = kitLevel.NewFilter(base, kitLevel.AllowDebug())
	default:
		filtered = kitLevel.NewFilter(base, kitLevel.AllowInfo())
	}

	return filtered
}

// Logger wraps a go-kit Logger.
type Logger struct {
	Level  Level
	base   kitLog.Logger
	logger *kitLog.SwapLogger
}

// NewLogger creates a new logger.
func NewLogger(opts Options) *Logger {
	level := stringToLevel(opts.Level)
	base := createBaseLogger(opts)
	filtered := createFilteredLogger(base, level)

	logger := new(kitLog.SwapLogger)
	logger.Swap(filtered)

	return &Logger{
		Level:  level,
		base:   base,
		logger: logger,
	}
}

// NewVoidLogger creates a void logger for testing purposes.
func NewVoidLogger() *Logger {
	nop := kitLog.NewNopLogger()

	logger := new(kitLog.SwapLogger)
	logger.Swap(nop)

	return &Logger{
		base:   nop,
		logger: logger,
	}
}

// With returns a new logger that always logs a set of key-value pairs (context).
func (l *Logger) With(kv ...interface{}) *Logger {
	level := l.Level
	base := kitLog.With(l.base, kv...)
	filtered := createFilteredLogger(base, level)

	logger := new(kitLog.SwapLogger)
	logger.Swap(filtered)

	return &Logger{
		Level:  level,
		base:   base,
		logger: logger,
	}
}

// SetLevel changes the level of logger.
func (l *Logger) SetLevel(level string) {
	l.Level = stringToLevel(level)
	l.logger.Swap(createFilteredLogger(l.base, l.Level))
}

// SetOptions resets a logger with new options.
func (l *Logger) SetOptions(opts Options) {
	l.Level = stringToLevel(opts.Level)
	l.base = createBaseLogger(opts)
	l.logger.Swap(createFilteredLogger(l.base, l.Level))
}

// Debugf logs a message in debug level.
func (l *Logger) Debugf(format string, v ...interface{}) {
	_ = kitLevel.Debug(l.logger).Log("message", fmt.Sprintf(format, v...))
}

// DebugKV logs key-value pairs in debug level.
func (l *Logger) DebugKV(kv ...interface{}) {
	_ = kitLevel.Debug(l.logger).Log(kv...)
}

// Infof logs a message in info level.
func (l *Logger) Infof(format string, v ...interface{}) {
	_ = kitLevel.Info(l.logger).Log("message", fmt.Sprintf(format, v...))
}

// InfoKV logs key-value pairs in info level.
func (l *Logger) InfoKV(kv ...interface{}) {
	_ = kitLevel.Info(l.logger).Log(kv...)
}

// Warnf logs a message in warn level.
func (l *Logger) Warnf(format string, v ...interface{}) {
	_ = kitLevel.Warn(l.logger).Log("message", fmt.Sprintf(format, v...))
}

// WarnKV logs key-value pairs in warn level.
func (l *Logger) WarnKV(kv ...interface{}) {
	_ = kitLevel.Warn(l.logger).Log(kv...)
}

// Errorf logs a message in error level.
func (l *Logger) Errorf(format string, v ...interface{}) {
	_ = kitLevel.Error(l.logger).Log("message", fmt.Sprintf(format, v...))
}

// ErrorKV logs key-value pairs in error level.
func (l *Logger) ErrorKV(kv ...interface{}) {
	_ = kitLevel.Error(l.logger).Log(kv...)
}

// The singleton logger.
var singleton = NewLogger(Options{
	Name:        "singleton",
	callerDepth: singletonCallerDepth,
})

// SetLevel changes the level of singleton logger.
func SetLevel(level string) {
	singleton.SetLevel(level)
}

// SetOptions set optional options for singleton logger.
func SetOptions(opts Options) {
	opts.callerDepth = 8
	singleton.SetOptions(opts)
}

// Debugf logs a message in debug level using singleton logger.
func Debugf(format string, v ...interface{}) {
	singleton.Debugf(format, v...)
}

// DebugKV logs key-value pairs in debug level using singleton logger.
func DebugKV(kv ...interface{}) {
	singleton.DebugKV(kv...)
}

// Infof logs a message in info level using singleton logger.
func Infof(format string, v ...interface{}) {
	singleton.Infof(format, v...)
}

// InfoKV logs key-value pairs in info level using singleton logger.
func InfoKV(kv ...interface{}) {
	singleton.InfoKV(kv...)
}

// Warnf logs a message in warn level using singleton logger.
func Warnf(format string, v ...interface{}) {
	singleton.Warnf(format, v...)
}

// WarnKV logs key-value pairs in warn level using singleton logger.
func WarnKV(kv ...interface{}) {
	singleton.WarnKV(kv...)
}

// Errorf logs a message in error level using singleton logger.
func Errorf(format string, v ...interface{}) {
	singleton.Errorf(format, v...)
}

// ErrorKV logs key-value pairs in error level using singleton logger.
func ErrorKV(kv ...interface{}) {
	singleton.ErrorKV(kv...)
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
