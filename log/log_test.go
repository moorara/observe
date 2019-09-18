package log

import (
	"bytes"
	"context"
	"errors"
	"testing"

	kitLog "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

type mockKitLogger struct {
	LogInKV     []interface{}
	LogOutError error
}

func (m *mockKitLogger) Log(kv ...interface{}) error {
	m.LogInKV = kv
	return m.LogOutError
}

func TestStringToLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel Level
	}{
		{"None", "none", NoneLevel},
		{"Error", "error", ErrorLevel},
		{"Warn", "warn", WarnLevel},
		{"Info", "info", InfoLevel},
		{"Debug", "debug", DebugLevel},
		{"Default", "trace", InfoLevel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			level := stringToLevel(tc.level)

			assert.Equal(t, tc.expectedLevel, level)
		})
	}
}

func TestCreateKitLogger(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{
			"NoOption",
			Options{},
		},
		{
			"WithName",
			Options{
				Name: "test",
			},
		},
		{
			"WithInfo",
			Options{
				Name:        "test",
				Environment: "local",
				Region:      "local",
			},
		},
		{
			"NoneLevel",
			Options{
				Level: "none",
			},
		},
		{
			"ErrorLevel",
			Options{
				Level: "error",
			},
		},
		{
			"WarnLevel",
			Options{
				Level: "warn",
			},
		},
		{
			"InfoLevel",
			Options{
				Level: "info",
			},
		},
		{
			"DebugLevel",
			Options{
				Level: "debug",
			},
		},
		{
			"JSONLogger",
			Options{
				Format: JSON,
			},
		},
		{
			"LogfmtLogger",
			Options{
				Format: Logfmt,
			},
		},
		{
			"WithWriter",
			Options{
				Writer: &bytes.Buffer{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kitLogger := createKitLogger(tc.opts)

			assert.NotNil(t, kitLogger)
		})
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name          string
		opts          Options
		expectedLevel Level
	}{
		{
			"NoOption",
			Options{},
			InfoLevel,
		},
		{
			"WithName",
			Options{
				Name: "test",
			},
			InfoLevel,
		},
		{
			"WithInfo",
			Options{
				Name:        "test",
				Environment: "local",
				Region:      "local",
			},
			InfoLevel,
		},
		{
			"NoneLevel",
			Options{
				Name:  "test",
				Level: "none",
			},
			NoneLevel,
		},
		{
			"ErrorLevel",
			Options{
				Name:  "test",
				Level: "error",
			},
			ErrorLevel,
		},
		{
			"WarnLevel",
			Options{
				Name:  "test",
				Level: "warn",
			},
			WarnLevel,
		},
		{
			"InfoLevel",
			Options{
				Name:  "test",
				Level: "info",
			},
			InfoLevel,
		},
		{
			"DebugLevel",
			Options{
				Name:  "test",
				Level: "debug",
			},
			DebugLevel,
		},
		{
			"JSONLogger",
			Options{
				Format: JSON,
			},
			InfoLevel,
		},
		{
			"LogfmtLogger",
			Options{
				Format: Logfmt,
			},
			InfoLevel,
		},
		{
			"WithWriter",
			Options{
				Writer: &bytes.Buffer{},
			},
			InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(T *testing.T) {
			logger := NewLogger(tc.opts)

			assert.NotNil(t, logger)
			assert.NotNil(t, logger.Logger)
			assert.Equal(t, logger.Level, tc.expectedLevel)
		})
	}
}

func TestNewVoidLogger(t *testing.T) {
	logger := NewVoidLogger()
	assert.NotNil(t, logger)
}

func TestLoggerWith(t *testing.T) {
	tests := []struct {
		logger *Logger
		kv     []interface{}
	}{
		{
			&Logger{
				Level:  InfoLevel,
				Logger: &kitLog.SwapLogger{},
			},
			[]interface{}{"version", "0.1.0", "revision", "1234567", "context", "test"},
		},
	}

	for _, tc := range tests {
		logger := tc.logger.With(tc.kv...)

		assert.NotNil(t, logger)
		assert.Equal(t, tc.logger.Level, logger.Level)
	}
}

func TestSetOptions(t *testing.T) {
	tests := []struct {
		name   string
		logger *Logger
		opts   Options
	}{
		{
			"NoOption",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{},
		},
		{
			"WithName",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Name: "test",
			},
		},
		{
			"WithInfo",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Name:        "test",
				Environment: "local",
				Region:      "local",
			},
		},
		{
			"NoneLevel",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "none",
			},
		},
		{
			"ErrorLevel",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "error",
			},
		},
		{
			"WarnLevel",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "warn",
			},
		},
		{
			"InfoLevel",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "info",
			},
		},
		{
			"DebugLevel",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "debug",
			},
		},
		{
			"JSONLogger",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Format: JSON,
			},
		},
		{
			"LogfmtLogger",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Format: Logfmt,
			},
		},
		{
			"WithWriter",
			&Logger{
				Logger: &kitLog.SwapLogger{},
			},
			Options{
				Writer: &bytes.Buffer{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.logger.SetOptions(tc.opts)
		})
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name          string
		mockKitLogger *mockKitLogger
		kv            []interface{}
		expectedError error
		expectedKV    []interface{}
	}{
		{
			"Error",
			&mockKitLogger{
				LogOutError: errors.New("log error"),
			},
			[]interface{}{"message", "operation failed", "operation", "test"},
			errors.New("log error"),
			[]interface{}{"message", "operation failed", "operation", "test"},
		},
		{
			"Success",
			&mockKitLogger{},
			[]interface{}{"message", "operation succeeded", "operation", "test"},
			nil,
			[]interface{}{"message", "operation succeeded", "operation", "test"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := &Logger{Logger: &kitLog.SwapLogger{}}
			logger.Logger.Swap(tc.mockKitLogger)

			t.Run("DebugLevel", func(t *testing.T) {
				logger.Debug(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("InfoLevel", func(t *testing.T) {
				logger.Info(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("WarnLevel", func(t *testing.T) {
				logger.Warn(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("ErrorLevel", func(t *testing.T) {
				logger.Error(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})
		})
	}
}

func TestSingletonSetOptions(t *testing.T) {
	tests := []struct {
		name          string
		opts          Options
		expectedLevel Level
	}{
		{
			"NoOption",
			Options{},
			InfoLevel,
		},
		{
			"WithName",
			Options{
				Name: "test",
			},
			InfoLevel,
		},
		{
			"WithInfo",
			Options{
				Name:        "test",
				Environment: "local",
				Region:      "local",
			},
			InfoLevel,
		},
		{
			"NoneLevel",
			Options{
				Name:  "test",
				Level: "none",
			},
			NoneLevel,
		},
		{
			"ErrorLevel",
			Options{
				Name:  "test",
				Level: "error",
			},
			ErrorLevel,
		},
		{
			"WarnLevel",
			Options{
				Name:  "test",
				Level: "warn",
			},
			WarnLevel,
		},
		{
			"InfoLevel",
			Options{
				Name:  "test",
				Level: "info",
			},
			InfoLevel,
		},
		{
			"DebugLevel",
			Options{
				Name:  "test",
				Level: "debug",
			},
			DebugLevel,
		},
		{
			"JSONLogger",
			Options{
				Format: JSON,
			},
			InfoLevel,
		},
		{
			"LogfmtLogger",
			Options{
				Format: Logfmt,
			},
			InfoLevel,
		},
		{
			"WithWriter",
			Options{
				Writer: &bytes.Buffer{},
			},
			InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SetOptions(tc.opts)

			assert.NotNil(t, singleton.Logger)
			assert.Equal(t, singleton.Level, tc.expectedLevel)
		})
	}
}

func TestSingletonLogger(t *testing.T) {
	tests := []struct {
		name          string
		mockKitLogger *mockKitLogger
		kv            []interface{}
		expectedError error
		expectedKV    []interface{}
	}{
		{
			"Error",
			&mockKitLogger{
				LogOutError: errors.New("log error"),
			},
			[]interface{}{"message", "operation failed", "operation", "test"},
			errors.New("log error"),
			[]interface{}{"message", "operation failed", "operation", "test"},
		},
		{
			"Success",
			&mockKitLogger{},
			[]interface{}{"message", "operation succeeded", "operation", "test"},
			nil,
			[]interface{}{"message", "operation succeeded", "operation", "test"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			singleton.Logger.Swap(tc.mockKitLogger)

			t.Run("DebugLevel", func(t *testing.T) {
				Debug(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("InfoLevel", func(t *testing.T) {
				Info(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("WarnLevel", func(t *testing.T) {
				Warn(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("ErrorLevel", func(t *testing.T) {
				Error(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})
		})
	}
}

func TestContextWithLogger(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		logger *Logger
	}{
		{
			name:   "NoLogger",
			ctx:    context.Background(),
			logger: nil,
		},
		{
			name:   "WithLogger",
			ctx:    context.Background(),
			logger: NewVoidLogger(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ContextWithLogger(tc.ctx, tc.logger)

			logger, ok := ctx.Value(loggerContextKey).(*Logger)
			assert.True(t, ok)
			assert.Equal(t, tc.logger, logger)
		})
	}
}

func TestLoggerFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "NoLogger",
			ctx:  context.Background(),
		},
		{
			name: "WithLogger",
			ctx:  context.WithValue(context.Background(), loggerContextKey, NewVoidLogger()),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := LoggerFromContext(tc.ctx)

			assert.NotEmpty(t, logger)
		})
	}
}
