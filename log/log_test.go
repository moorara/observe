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
		{"Empty", "", InfoLevel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			level := stringToLevel(tc.level)
			assert.Equal(t, tc.expectedLevel, level)
		})
	}
}

func TestCreateBaseLogger(t *testing.T) {
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
			"WithContext",
			Options{
				Name:        "test",
				Environment: "local",
				Region:      "local",
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
			base := createBaseLogger(tc.opts)
			assert.NotNil(t, base)
		})
	}
}

func TestCreateFilteredLogger(t *testing.T) {
	tests := []struct {
		name  string
		base  kitLog.Logger
		level Level
	}{
		{
			"NoneLevel",
			kitLog.NewNopLogger(),
			NoneLevel,
		},
		{
			"ErrorLevel",
			kitLog.NewNopLogger(),
			ErrorLevel,
		},
		{
			"WarnLevel",
			kitLog.NewNopLogger(),
			WarnLevel,
		},
		{
			"InfoLevel",
			kitLog.NewNopLogger(),
			InfoLevel,
		},
		{
			"DebugLevel",
			kitLog.NewNopLogger(),
			DebugLevel,
		},
		{
			"InvalidLevel",
			kitLog.NewNopLogger(),
			Level(99),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filtered := createFilteredLogger(tc.base, tc.level)
			assert.NotNil(t, filtered)
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
			"WithContext",
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
			assert.NotNil(t, logger.base)
			assert.NotNil(t, logger.logger)
			assert.Equal(t, tc.expectedLevel, logger.Level)
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
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			[]interface{}{
				"version", "0.1.0",
				"revision", "1234567",
				"context", "test",
			},
		},
	}

	for _, tc := range tests {
		logger := tc.logger.With(tc.kv...)

		assert.NotNil(t, logger)
		assert.Equal(t, tc.logger.Level, logger.Level)
	}
}

func TestLoggerSetLevel(t *testing.T) {
	tests := []struct {
		name          string
		logger        *Logger
		level         string
		expectedLevel Level
	}{
		{
			"NoneLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			"none",
			NoneLevel,
		},
		{
			"ErrorLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			"error",
			ErrorLevel,
		},
		{
			"WarnLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			"warn",
			WarnLevel,
		},
		{
			"InfoLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			"info",
			InfoLevel,
		},
		{
			"DebugLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			"debug",
			DebugLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.logger.SetLevel(tc.level)

			assert.NotNil(t, tc.logger.logger)
			assert.Equal(t, tc.expectedLevel, tc.logger.Level)
		})
	}
}

func TestLoggerSetOptions(t *testing.T) {
	tests := []struct {
		name          string
		logger        *Logger
		opts          Options
		expectedLevel Level
	}{
		{
			"NoOption",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{},
			InfoLevel,
		},
		{
			"WithName",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Name: "test",
			},
			InfoLevel,
		},
		{
			"WithContext",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Name:        "test",
				Environment: "local",
				Region:      "local",
			},
			InfoLevel,
		},
		{
			"NoneLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "none",
			},
			NoneLevel,
		},
		{
			"ErrorLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "error",
			},
			ErrorLevel,
		},
		{
			"WarnLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "warn",
			},
			WarnLevel,
		},
		{
			"InfoLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "info",
			},
			InfoLevel,
		},
		{
			"DebugLevel",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Level: "debug",
			},
			DebugLevel,
		},
		{
			"JSONLogger",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Format: JSON,
			},
			InfoLevel,
		},
		{
			"LogfmtLogger",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Format: Logfmt,
			},
			InfoLevel,
		},
		{
			"WithWriter",
			&Logger{
				base:   kitLog.NewNopLogger(),
				logger: &kitLog.SwapLogger{},
			},
			Options{
				Writer: &bytes.Buffer{},
			},
			InfoLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.logger.SetOptions(tc.opts)

			assert.NotNil(t, tc.logger.base)
			assert.NotNil(t, tc.logger.logger)
			assert.Equal(t, tc.expectedLevel, tc.logger.Level)
		})
	}
}

func TestLoggerMessage(t *testing.T) {
	tests := []struct {
		name          string
		mockKitLogger *mockKitLogger
		format        string
		vals          []interface{}
		expectedError error
		expectedKV    []interface{}
	}{
		{
			"Error",
			&mockKitLogger{
				LogOutError: errors.New("log error"),
			},
			"operation failed: %s", []interface{}{"no capacity"},
			errors.New("log error"),
			[]interface{}{"message", "operation failed: no capacity"},
		},
		{
			"Success",
			&mockKitLogger{},
			"operation succeeded: %s", []interface{}{"test"},
			nil,
			[]interface{}{"message", "operation succeeded: test"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := &Logger{logger: &kitLog.SwapLogger{}}
			logger.logger.Swap(tc.mockKitLogger)

			t.Run("Debugf", func(t *testing.T) {
				logger.Debugf(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("Infof", func(t *testing.T) {
				logger.Infof(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("Warnf", func(t *testing.T) {
				logger.Warnf(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("Errorf", func(t *testing.T) {
				logger.Errorf(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})
		})
	}
}

func TestLoggerKV(t *testing.T) {
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
			[]interface{}{"message", "operation failed", "reason", "no capacity"},
			errors.New("log error"),
			[]interface{}{"message", "operation failed", "reason", "no capacity"},
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
			logger := &Logger{logger: &kitLog.SwapLogger{}}
			logger.logger.Swap(tc.mockKitLogger)

			t.Run("DebugKV", func(t *testing.T) {
				logger.DebugKV(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("InfoKV", func(t *testing.T) {
				logger.InfoKV(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("WarnKV", func(t *testing.T) {
				logger.WarnKV(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("ErrorKV", func(t *testing.T) {
				logger.ErrorKV(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})
		})
	}
}

func TestSingletonSetLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		expectedLevel Level
	}{
		{
			"NoneLevel",
			"none",
			NoneLevel,
		},
		{
			"ErrorLevel",
			"error",
			ErrorLevel,
		},
		{
			"WarnLevel",
			"warn",
			WarnLevel,
		},
		{
			"InfoLevel",
			"info",
			InfoLevel,
		},
		{
			"DebugLevel",
			"debug",
			DebugLevel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SetLevel(tc.level)

			assert.NotNil(t, singleton.logger)
			assert.Equal(t, tc.expectedLevel, singleton.Level)
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
			"WithContext",
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
				Level: "none",
			},
			NoneLevel,
		},
		{
			"ErrorLevel",
			Options{
				Level: "error",
			},
			ErrorLevel,
		},
		{
			"WarnLevel",
			Options{
				Level: "warn",
			},
			WarnLevel,
		},
		{
			"InfoLevel",
			Options{
				Level: "info",
			},
			InfoLevel,
		},
		{
			"DebugLevel",
			Options{
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

			assert.NotNil(t, singleton.base)
			assert.NotNil(t, singleton.logger)
			assert.Equal(t, tc.expectedLevel, singleton.Level)
		})
	}
}

func TestSingletonLoggerMessage(t *testing.T) {
	tests := []struct {
		name          string
		mockKitLogger *mockKitLogger
		format        string
		vals          []interface{}
		expectedError error
		expectedKV    []interface{}
	}{
		{
			"Error",
			&mockKitLogger{
				LogOutError: errors.New("log error"),
			},
			"operation failed: %s", []interface{}{"no capacity"},
			errors.New("log error"),
			[]interface{}{"message", "operation failed: no capacity"},
		},
		{
			"Success",
			&mockKitLogger{},
			"operation succeeded: %s", []interface{}{"test"},
			nil,
			[]interface{}{"message", "operation succeeded: test"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			singleton.logger.Swap(tc.mockKitLogger)

			t.Run("Debugf", func(t *testing.T) {
				Debugf(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("Infof", func(t *testing.T) {
				Infof(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("Warnf", func(t *testing.T) {
				Warnf(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("Errorf", func(t *testing.T) {
				Errorf(tc.format, tc.vals...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})
		})
	}
}

func TestSingletonLoggerKV(t *testing.T) {
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
			[]interface{}{"message", "operation failed", "reason", "no capacity"},
			errors.New("log error"),
			[]interface{}{"message", "operation failed", "reason", "no capacity"},
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
			singleton.logger.Swap(tc.mockKitLogger)

			t.Run("DebugKV", func(t *testing.T) {
				DebugKV(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("InfoKV", func(t *testing.T) {
				InfoKV(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("WarnKV", func(t *testing.T) {
				WarnKV(tc.kv...)
				for _, val := range tc.expectedKV {
					assert.Contains(t, tc.mockKitLogger.LogInKV, val)
				}
			})

			t.Run("ErrorKV", func(t *testing.T) {
				ErrorKV(tc.kv...)
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
