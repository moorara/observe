package observe

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the singleton logger.
var Logger *zap.Logger

// LoggerOptions are optional configurations for creating a logger.
// level can be "debug", "info", "warn", "error", or "none" (case-insensitive).
type LoggerOptions struct {
	Name        string
	Environment string
	Region      string
	Level       string
}

// NewZapLogger creates a new zap logger.
func NewZapLogger(opts LoggerOptions) (*zap.Logger, *zap.Config) {
	config := zap.NewProductionConfig()
	config.Encoding = "json"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.NameKey = "logger"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	config.OutputPaths = []string{"stdout"}
	config.InitialFields = make(map[string]interface{})

	if opts.Name != "" {
		config.InitialFields["logger"] = opts.Name
	}

	if opts.Environment != "" {
		config.InitialFields["environment"] = opts.Environment
	}

	if opts.Region != "" {
		config.InitialFields["region"] = opts.Region
	}

	switch strings.ToLower(opts.Level) {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "": // default
		fallthrough
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "none":
		fallthrough
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.Level(99))
	}

	logger, _ := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(0),
	)

	return logger, &config
}

// SetLogger sets the singleton logger.
// It needs to be called once to initialize the singleton logger.
func SetLogger(logger *zap.Logger) {
	Logger = logger
}
