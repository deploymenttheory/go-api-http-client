// zaplogger_logger_test.go
package logger

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestDefaultLogger_SetLevel tests the SetLevel method of defaultLogger
func TestDefaultLogger_SetLevel(t *testing.T) {
	logger := zap.NewNop()
	dLogger := &defaultLogger{logger: logger}

	dLogger.SetLevel(LogLevelWarn)
	assert.Equal(t, LogLevelWarn, dLogger.GetLogLevel())
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

// TestDefaultLogger_GetLogLevel verifies that the GetLogLevel method of the defaultLogger struct
// accurately returns the logger's current log level setting. This test ensures that the log level
// set within the defaultLogger is properly retrievable.
func TestDefaultLogger_GetLogLevel(t *testing.T) {
	// Define test cases for each log level
	logLevels := []struct {
		level    LogLevel
		expected LogLevel
	}{
		{LogLevelDebug, LogLevelDebug},
		{LogLevelInfo, LogLevelInfo},
		{LogLevelWarn, LogLevelWarn},
		{LogLevelError, LogLevelError},
		{LogLevelDPanic, LogLevelDPanic},
		{LogLevelPanic, LogLevelPanic},
		{LogLevelFatal, LogLevelFatal},
	}

	for _, tc := range logLevels {
		t.Run(fmt.Sprintf("LogLevel %d", tc.level), func(t *testing.T) {
			dLogger := &defaultLogger{logLevel: tc.level}

			// Assert that GetLogLevel returns the correct log level for each case
			assert.Equal(t, tc.expected, dLogger.GetLogLevel(), fmt.Sprintf("GetLogLevel should return %d for set log level %d", tc.expected, tc.level))
		})
	}
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

			// Check if the logger's log level matches the expected log level
			assert.Equal(t, logger.Core().Enabled(zapcore.Level(tt.expectedLevel.Level())), true, "Logger level should match expected")
		})
	}
}
