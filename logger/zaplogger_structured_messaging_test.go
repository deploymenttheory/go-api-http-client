// zaplogger_structured_messaging_test.go
package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestParseLogLevelFromString tests the conversion from string to LogLevel
func TestParseLogLevelFromString(t *testing.T) {
	tests := []struct {
		levelStr      string
		expectedLevel LogLevel
	}{
		{"LogLevelDebug", LogLevelDebug},
		{"LogLevelInfo", LogLevelInfo},
		{"LogLevelWarn", LogLevelWarn},
		{"LogLevelError", LogLevelError},
		{"LogLevelDPanic", LogLevelDPanic},
		{"LogLevelPanic", LogLevelPanic},
		{"LogLevelFatal", LogLevelFatal},
		{"Invalid", LogLevelNone},
	}

	for _, tt := range tests {
		t.Run(tt.levelStr, func(t *testing.T) {
			result := ParseLogLevelFromString(tt.levelStr)
			assert.Equal(t, tt.expectedLevel, result)
		})
	}
}

// TestDefaultLogger_SetLevel tests the SetLevel method of defaultLogger
func TestDefaultLogger_SetLevel(t *testing.T) {
	logger := zap.NewNop()
	dLogger := &defaultLogger{logger: logger}

	dLogger.SetLevel(LogLevelWarn)
	assert.Equal(t, LogLevelWarn, dLogger.GetLogLevel())
}

// TestDefaultLogger_LoggingMethods tests the logging methods (Debug, Info, Warn, Error, Panic, Fatal)
// Note: Actual logging output is not tested here due to the complexity of capturing log output
func TestDefaultLogger_LoggingMethods(t *testing.T) {
	logger := zap.NewNop() // Using zap's No-op logger for testing
	dLogger := &defaultLogger{logger: logger, logLevel: LogLevelDebug}

	// Only testing method calls and log level checks, not the actual logging output
	assert.NotPanics(t, func() { dLogger.Debug("test message") }, "Debug should not panic")
	assert.NotPanics(t, func() { dLogger.Info("test message") }, "Info should not panic")
	assert.NotPanics(t, func() { dLogger.Warn("test message") }, "Warn should not panic")
	assert.NoError(t, dLogger.Error("test message"), "Error should not return an error")
	assert.NotPanics(t, func() { dLogger.Panic("test message") }, "Panic should not panic in this test context")
	assert.NotPanics(t, func() { dLogger.Fatal("test message") }, "Fatal should not panic in this test context")
}

// TestDefaultLogger_With tests the With method functionality
func TestDefaultLogger_With(t *testing.T) {
	logger := zap.NewNop()
	dLogger := &defaultLogger{logger: logger, logLevel: LogLevelInfo}

	newLogger := dLogger.With(zap.String("key", "value"))
	assert.NotNil(t, newLogger, "New logger should not be nil")

	// Assert that newLogger is a *defaultLogger and has a modified zap.Logger
	assert.IsType(t, &defaultLogger{}, newLogger, "New logger should be of type *defaultLogger")
}

// TestGetLoggerBasedOnEnv tests the GetLoggerBasedOnEnv function for different environment settings
func TestGetLoggerBasedOnEnv(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedLevel zap.AtomicLevel
	}{
		{"DevelopmentLogger", "development", zap.NewAtomicLevelAt(zap.DebugLevel)},
		{"ProductionLogger", "production", zap.NewAtomicLevelAt(zap.InfoLevel)},
		{"DefaultToProduction", "", zap.NewAtomicLevelAt(zap.InfoLevel)}, // default case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set APP_ENV to the desired test value
			os.Setenv("APP_ENV", tt.envValue)
			defer os.Unsetenv("APP_ENV") // Clean up

			logger := GetLoggerBasedOnEnv()

			// Since we cannot directly access the logger's level, we check the logger's development/production status
			// which indirectly tells us about the log level configuration
			cfg := zap.NewProductionConfig()
			if tt.envValue == "development" {
				cfg = zap.NewDevelopmentConfig()
			}

			assert.Equal(t, cfg.Level.Level(), logger.Core().Enabled(zapcore.Level(tt.expectedLevel.Level())), "Logger level should match expected")
		})
	}
}
