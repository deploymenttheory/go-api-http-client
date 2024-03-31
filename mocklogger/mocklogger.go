package mocklogger

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger" // Assuming this is the package where Logger interface is defined
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MockLogger is a mock type for the Logger interface, embedding a *zap.Logger to satisfy the type requirement.
type MockLogger struct {
	mock.Mock
	*zap.Logger
	logLevel logger.LogLevel
}

// NewMockLogger creates a new instance of MockLogger with an embedded no-op *zap.Logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Logger: zap.NewNop(),
	}
}

// Ensure MockLogger implements the logger.Logger interface from the logger package
var _ logger.Logger = (*MockLogger)(nil)

func (m *MockLogger) GetLogLevel() logger.LogLevel {
	args := m.Called()
	if args.Get(0) != nil { // Check if the Called method has a return value
		return args.Get(0).(logger.LogLevel)
	}
	return logger.LogLevelNone // Return LogLevelNone if no specific log level is set
}

// SetLevel sets the logging level of the MockLogger.
// This controls the verbosity of the logger, allowing it to filter out logs below the set level.
func (m *MockLogger) SetLevel(level logger.LogLevel) {
	m.logLevel = level
	m.Called(level)
}

// With adds contextual key-value pairs to the MockLogger and returns a new logger instance with this context.
// This is useful for adding common fields to all subsequent logs produced by the logger.
func (m *MockLogger) With(fields ...zapcore.Field) logger.Logger {
	m.Called(fields)

	return m
}

// Debug logs a message at the Debug level.
func (m *MockLogger) Debug(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
	if m.logLevel <= logger.LogLevelDebug {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

// Info logs a message at the Info level.
func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
	if m.logLevel <= logger.LogLevelInfo {
		fmt.Printf("[INFO] %s\n", msg)
	}
}

// Error logs a message at the Error level and returns an error.
func (m *MockLogger) Error(msg string, fields ...zapcore.Field) error {
	m.Called(msg, fields)
	if m.logLevel <= logger.LogLevelError {
		fmt.Printf("[ERROR] %s\n", msg)
	}
	return errors.New(msg)
}

// Warn logs a message at the Warn level.
func (m *MockLogger) Warn(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
	if m.logLevel <= logger.LogLevelWarn {
		fmt.Printf("[WARN] %s\n", msg)
	}
}

// Panic logs a message at the Panic level and then panics.
func (m *MockLogger) Panic(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
	if m.logLevel <= logger.LogLevelPanic {
		fmt.Printf("[PANIC] %s\n", msg)
		panic(msg)
	}
}

// Fatal logs a message at the Fatal level and then calls os.Exit(1).
func (m *MockLogger) Fatal(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
	if m.logLevel <= logger.LogLevelFatal {
		fmt.Printf("[FATAL] %s\n", msg)
		os.Exit(1)
	}
}

// LogRequestStart logs the start of an HTTP request.
func (m *MockLogger) LogRequestStart(event string, requestID string, userID string, method string, url string, headers map[string][]string) {
	m.Called(event, requestID, userID, method, url, headers)
	// Mock logging implementation...
}

// LogRequestEnd logs the end of an HTTP request.
func (m *MockLogger) LogRequestEnd(event string, method string, url string, statusCode int, duration time.Duration) {
	m.Called(event, method, url, statusCode, duration)
	// Mock logging implementation...
}

// LogError logs an error event.
func (m *MockLogger) LogError(event string, method string, url string, statusCode int, serverStatusMessage string, err error, rawResponse string) {
	m.Called(event, method, url, statusCode, serverStatusMessage, err, rawResponse)
	// Mock logging implementation...
}

// Example for LogAuthTokenError:
func (m *MockLogger) LogAuthTokenError(event string, method string, url string, statusCode int, err error) {
	m.Called(event, method, url, statusCode, err)
	// Mock logging implementation...
}

// LogCookies logs information about cookies.
func (m *MockLogger) LogCookies(direction string, obj interface{}, method, url string) {
	// Use the mock framework to record that LogCookies was called with the specified arguments
	m.Called(direction, obj, method, url)
	fmt.Printf("[COOKIES] Direction: %s, Object: %v, Method: %s, URL: %s\n", direction, obj, method, url)
}

// LogRetryAttempt logs a retry attempt.
func (m *MockLogger) LogRetryAttempt(event string, method string, url string, attempt int, reason string, waitDuration time.Duration, err error) {
	m.Called(event, method, url, attempt, reason, waitDuration, err)
	// Mock logging implementation...
	fmt.Printf("[RETRY ATTEMPT] Event: %s, Method: %s, URL: %s, Attempt: %d, Reason: %s, Wait Duration: %s, Error: %v\n", event, method, url, attempt, reason, waitDuration, err)
}

// LogRateLimiting logs rate limiting events.
func (m *MockLogger) LogRateLimiting(event string, method string, url string, retryAfter string, waitDuration time.Duration) {
	m.Called(event, method, url, retryAfter, waitDuration)
	// Mock logging implementation...
	fmt.Printf("[RATE LIMITING] Event: %s, Method: %s, URL: %s, Retry After: %s, Wait Duration: %s\n", event, method, url, retryAfter, waitDuration)
}

// LogResponse logs HTTP responses.
func (m *MockLogger) LogResponse(event string, method string, url string, statusCode int, responseBody string, responseHeaders map[string][]string, duration time.Duration) {
	m.Called(event, method, url, statusCode, responseBody, responseHeaders, duration)
	// Mock logging implementation...
	fmt.Printf("[RESPONSE] Event: %s, Method: %s, URL: %s, Status Code: %d, Response Body: %s, Response Headers: %v, Duration: %s\n", event, method, url, statusCode, responseBody, responseHeaders, duration)
}
