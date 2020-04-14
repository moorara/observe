package observe

import (
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestNewZapLogger(t *testing.T) {
	tests := []struct {
		name             string
		opts             LoggerOptions
		expectedLevel    zapcore.Level
		expectedEncoding string
	}{
		{
			name:             "NoOption",
			opts:             LoggerOptions{},
			expectedLevel:    zapcore.InfoLevel,
			expectedEncoding: "json",
		},
		{
			name: "WithMetadata",
			opts: LoggerOptions{
				Name:        "test",
				Environment: "local",
				Region:      "local",
			},
			expectedLevel:    zapcore.InfoLevel,
			expectedEncoding: "json",
		},
		{
			name: "LevelDebug",
			opts: LoggerOptions{
				Name:  "test",
				Level: "debug",
			},
			expectedLevel:    zapcore.DebugLevel,
			expectedEncoding: "json",
		},
		{
			name: "LevelInfo",
			opts: LoggerOptions{
				Name:  "test",
				Level: "info",
			},
			expectedLevel:    zapcore.InfoLevel,
			expectedEncoding: "json",
		},
		{
			name: "LevelWarn",
			opts: LoggerOptions{
				Name:  "test",
				Level: "warn",
			},
			expectedLevel:    zapcore.WarnLevel,
			expectedEncoding: "json",
		},
		{
			name: "LevelError",
			opts: LoggerOptions{
				Name:  "test",
				Level: "error",
			},
			expectedLevel:    zapcore.ErrorLevel,
			expectedEncoding: "json",
		},
		{
			name: "LevelNone",
			opts: LoggerOptions{
				Name:  "test",
				Level: "none",
			},
			expectedLevel:    zapcore.Level(99),
			expectedEncoding: "json",
		},
		{
			name: "InvalidLevel",
			opts: LoggerOptions{
				Name:  "test",
				Level: "invalid",
			},
			expectedLevel:    zapcore.Level(99),
			expectedEncoding: "json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(T *testing.T) {
			logger, config := NewZapLogger(tc.opts)

			assert.NotNil(t, logger)
			assert.NotNil(t, config)
			assert.Equal(t, tc.expectedLevel, config.Level.Level())
			assert.Equal(t, tc.expectedEncoding, config.Encoding)
		})
	}
}

func TestSetLogger(t *testing.T) {
	tests := []struct {
		name   string
		logger *zap.Logger
	}{
		{
			name:   "OK",
			logger: zap.NewExample(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SetLogger(tc.logger)

			assert.Equal(t, tc.logger, Logger)
		})
	}
}
