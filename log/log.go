package log

import (
	"io"
	"os"
	"strings"

	kitLog "github.com/go-kit/kit/log"
	kitLevel "github.com/go-kit/kit/log/level"
)

// Format is the type for output format
type Format int

const (
	// JSON represents a json logger
	JSON Format = iota
	// Logfmt represents logfmt logger
	Logfmt
)

// Level is the type for logging level
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

// Options contains optional options for Logger
type Options struct {
	depth       int
	Name        string
	Environment string
	Region      string
	Level       string
	Format      Format
	Writer      io.Writer
}

func createKitLogger(opts Options) kitLog.Logger {
	var logger kitLog.Logger

	if opts.depth == 0 {
		opts.depth = 6
	}

	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	switch opts.Format {
	case Logfmt:
		logger = kitLog.NewLogfmtLogger(opts.Writer)
	case JSON:
		fallthrough
	default:
		logger = kitLog.NewJSONLogger(opts.Writer)
	}

	// This is not required since SwapLogger can be used concurrently
	// logger = kitLog.NewSyncLogger(logger)

	logger = kitLog.With(logger,
		"caller", kitLog.Caller(opts.depth),
		"timestamp", kitLog.DefaultTimestampUTC,
	)

	if opts.Name != "" {
		logger = kitLog.With(logger, "logger", opts.Name)
	}

	if opts.Environment != "" {
		logger = kitLog.With(logger, "environment", opts.Environment)
	}

	if opts.Region != "" {
		logger = kitLog.With(logger, "region", opts.Region)
	}

	switch strings.ToLower(opts.Level) {
	case "none":
		logger = kitLevel.NewFilter(logger, kitLevel.AllowNone())
	case "error":
		logger = kitLevel.NewFilter(logger, kitLevel.AllowError())
	case "warn":
		logger = kitLevel.NewFilter(logger, kitLevel.AllowWarn())
	case "info":
		logger = kitLevel.NewFilter(logger, kitLevel.AllowInfo())
	case "debug":
		logger = kitLevel.NewFilter(logger, kitLevel.AllowDebug())
	default:
		logger = kitLevel.NewFilter(logger, kitLevel.AllowInfo())
	}

	return logger
}

// Logger wraps a go-kit Logger
type Logger struct {
	Level  Level
	Logger *kitLog.SwapLogger
}

func (l *Logger) swap(logger kitLog.Logger) {
	l.Logger.Swap(logger)
}

// NewLogger creates a new logger
func NewLogger(opts Options) *Logger {
	logger := &Logger{
		Level:  stringToLevel(opts.Level),
		Logger: new(kitLog.SwapLogger),
	}

	kitLogger := createKitLogger(opts)
	logger.swap(kitLogger)

	return logger
}

// NewVoidLogger creates a void logger for testing purposes
func NewVoidLogger() *Logger {
	logger := &Logger{
		Logger: new(kitLog.SwapLogger),
	}

	kitLogger := kitLog.NewNopLogger()
	logger.swap(kitLogger)

	return logger
}

// With returns a new logger that always logs a set of key-value pairs (context)
func (l *Logger) With(kv ...interface{}) *Logger {
	logger := &Logger{
		Level:  l.Level,
		Logger: new(kitLog.SwapLogger),
	}

	kitLogger := kitLog.With(l.Logger, kv...)
	logger.swap(kitLogger)

	return logger
}

// SetOptions resets a logger with new options
func (l *Logger) SetOptions(opts Options) {
	kitLogger := createKitLogger(opts)
	l.swap(kitLogger)
}

// Debug logs in debug level
func (l *Logger) Debug(kv ...interface{}) error {
	return kitLevel.Debug(l.Logger).Log(kv...)
}

// Info logs in debug level
func (l *Logger) Info(kv ...interface{}) error {
	return kitLevel.Info(l.Logger).Log(kv...)
}

// Warn logs in debug level
func (l *Logger) Warn(kv ...interface{}) error {
	return kitLevel.Warn(l.Logger).Log(kv...)
}

// Error logs in debug level
func (l *Logger) Error(kv ...interface{}) error {
	return kitLevel.Error(l.Logger).Log(kv...)
}

// The singleton logger
var singleton = NewLogger(Options{
	depth: 7,
	Name:  "singleton",
})

// SetOptions set optional options for singleton logger
func SetOptions(opts Options) {
	opts.depth = 7
	singleton.SetOptions(opts)
	singleton.Level = stringToLevel(opts.Level)
}

// Debug logs a debug-level event using singleton logger
func Debug(kv ...interface{}) error {
	return singleton.Debug(kv...)
}

// Info logs an info-level event using singleton logger
func Info(kv ...interface{}) error {
	return singleton.Info(kv...)
}

// Warn logs a warn-level event using singleton logger
func Warn(kv ...interface{}) error {
	return singleton.Warn(kv...)
}

// Error logs an error-level event using singleton logger
func Error(kv ...interface{}) error {
	return singleton.Error(kv...)
}
