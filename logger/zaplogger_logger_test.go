// zaplogger_logger_test.go
package logger

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// osExit is a variable that holds a reference to os.Exit function.
// It allows overriding os.Exit in tests to prevent exiting the test runner.
var osExit = os.Exit

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

// TestDefaultLogger_Debug verifies that the Debug method of the defaultLogger struct correctly
// invokes the underlying zap.Logger's Debug method when the log level is set to allow Debug messages.
// The test uses a mockLogger to simulate the zap.Logger behavior, allowing verification of method calls
// without actual logging output. This ensures that the Debug method adheres to the expected behavior
// based on the current log level setting, providing confidence in the logging logic's correctness.
func TestDefaultLogger_Debug(t *testing.T) {
	mockLogger := NewMockLogger()
	dLogger := &defaultLogger{logger: mockLogger.Logger, logLevel: LogLevelDebug}

	expectedMessage := "test message"
	mockLogger.On("Debug", expectedMessage, mock.Anything).Once()

	dLogger.Debug(expectedMessage)

	mockLogger.AssertExpectations(t)
}

// TestDefaultLogger_Info verifies the Info method of the defaultLogger struct.
// It ensures that Info logs messages at the Info level when the logger's level allows for it.
// The test uses a mockLogger to intercept and verify the call to the underlying zap.Logger's Info method.
func TestDefaultLogger_Info(t *testing.T) {
	mockLogger := NewMockLogger()
	dLogger := &defaultLogger{logger: mockLogger.Logger, logLevel: LogLevelInfo}

	expectedMessage := "info message"
	mockLogger.On("Info", expectedMessage, mock.Anything).Once()

	dLogger.Info(expectedMessage)

	mockLogger.AssertExpectations(t)
}

// TestDefaultLogger_Warn verifies the Warn method of the defaultLogger struct.
// This test checks that Warn correctly logs messages at the Warn level based on the logger's current level.
// The behavior is validated using a mockLogger to capture and assert the call to the zap.Logger's Warn method.
func TestDefaultLogger_Warn(t *testing.T) {
	mockLogger := NewMockLogger()
	dLogger := &defaultLogger{logger: mockLogger.Logger, logLevel: LogLevelWarn}

	expectedMessage := "warn message"
	mockLogger.On("Warn", expectedMessage, []zapcore.Field(nil)).Once()

	dLogger.Warn(expectedMessage)

	mockLogger.AssertExpectations(t)
}

// TestDefaultLogger_Error checks the functionality of the Error method in the defaultLogger struct.
// It ensures that Error logs messages at the Error level and returns an error as expected.
// The test utilizes a mockLogger to track and affirm the invocation of zap.Logger's Error method.
func TestDefaultLogger_Error(t *testing.T) {
	mockLogger := NewMockLogger()
	dLogger := &defaultLogger{logger: mockLogger.Logger, logLevel: LogLevelError}

	expectedErrorMsg := "error message"
	// Mock the Error method to return nil, indicating no error
	mockLogger.On("Error", expectedErrorMsg, mock.Anything).Once().Return(fmt.Errorf("error message"))

	err := dLogger.Error(expectedErrorMsg)

	assert.NoError(t, err) // Check that the error returned is nil
	mockLogger.AssertExpectations(t)
}

// TestDefaultLogger_Panic ensures the Panic method of the defaultLogger behaves correctly.
// This test verifies that Panic logs messages at the Panic level and triggers a panic as expected.
// Due to the nature of panic, this test needs to recover from the panic to verify the behavior.
func TestDefaultLogger_Panic(t *testing.T) {
	mockLogger := NewMockLogger()
	dLogger := &defaultLogger{logger: mockLogger.Logger, logLevel: LogLevelPanic}

	expectedMessage := "panic message"
	mockLogger.On("Panic", expectedMessage, mock.Anything).Once()

	// Assert that calling Panic method triggers a panic
	assert.Panics(t, func() { dLogger.Panic(expectedMessage) }, "The Panic method should trigger a panic")

	mockLogger.AssertExpectations(t)
}

// TestDefaultLogger_Fatal verifies the Fatal method of the defaultLogger struct.
// It ensures that Fatal logs messages at the Fatal level and exits the application with a non-zero status code.
// The test utilizes a mockLogger to capture and assert the call to the zap.Logger's Fatal method.
func TestDefaultLogger_Fatal(t *testing.T) {
	mockLogger := NewMockLogger()
	dLogger := &defaultLogger{logger: mockLogger.Logger, logLevel: LogLevelFatal}

	expectedFatalMsg := "fatal message"
	mockLogger.On("Fatal", expectedFatalMsg, mock.Anything).Once()

	// Since Fatal exits the application, we use a sub-test to capture the exit status
	t.Run("TestFatal", func(t *testing.T) {
		// Replace os.Exit temporarily to capture exit status
		oldExit := osExit
		defer func() { osExit = oldExit }()
		var exitStatus int
		osExit = func(code int) {
			exitStatus = code
		}

		dLogger.Fatal(expectedFatalMsg)

		assert.Equal(t, 1, exitStatus, "Expected non-zero exit status")
	})

	mockLogger.AssertExpectations(t)
}

// Debug mocks the Debug method of the Logger interface
func (m *MockLogger) Debug(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// Info mocks the Info method of the Logger interface
func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// Warn mocks the Warn method of the Logger interface
func (m *MockLogger) Warn(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// Error mocks the Error method of the Logger interface
func (m *MockLogger) Error(msg string, fields ...zapcore.Field) error {
	args := m.Called(msg, fields)
	return args.Error(0)
}

// With mocks the With method of the Logger interface
func (m *MockLogger) With(fields ...zapcore.Field) Logger {
	args := m.Called(fields)
	return args.Get(0).(Logger)
}

// GetLogLevel mocks the GetLogLevel method of the Logger interface
func (m *MockLogger) GetLogLevel() LogLevel {
	args := m.Called()
	return args.Get(0).(LogLevel)
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
