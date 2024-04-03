// mocklogger/mocklogger.go
package mocklogger

import (
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger" // Assuming this is the package where Logger interface is defined
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

// GetLogLevel mocks the GetLogLevel method of the Logger interface.
func (m *MockLogger) GetLogLevel() logger.LogLevel {
	args := m.Called()
	return args.Get(0).(logger.LogLevel) // Assuming that the first argument is the log level
}

// SetLevel sets the logging level of the MockLogger.
// This controls the verbosity of the logger, allowing it to filter out logs below the set level.
func (m *MockLogger) SetLevel(level logger.LogLevel) {
	m.logLevel = level
	m.Called(level)
}

// With adds contextual key-value pairs to the MockLogger and returns a new logger instance with this context.
// This is useful for adding common fields to all subsequent logs produced by the logger.
func (m *MockLogger) With(fields ...zap.Field) logger.Logger {
	m.Called(fields)
	newMock := NewMockLogger()
	newMock.logLevel = m.logLevel
	return newMock
}

// Debug logs a message at the Debug level.
func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

// Info logs a message at the Info level.
func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

// Error logs a message at the Error level and returns an error.
func (m *MockLogger) Error(msg string, fields ...zap.Field) error {
	m.Called(msg, fields)
	return m.Called(msg).Error(0)
}

// Warn logs a message at the Warn level.
func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

// Panic logs a message at the Panic level and then panics.
func (m *MockLogger) Panic(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

// Fatal logs a message at the Fatal level and then calls os.Exit(1).
func (m *MockLogger) Fatal(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

// LogRequestStart logs the start of an HTTP request.
func (m *MockLogger) LogRequestStart(event string, requestID string, userID string, method string, url string, headers map[string][]string) {
	m.Called(event, requestID, userID, method, url, headers)
}

// LogRequestEnd logs the end of an HTTP request.
func (m *MockLogger) LogRequestEnd(event string, method string, url string, statusCode int, duration time.Duration) {
	m.Called(event, method, url, statusCode, duration)
}

// LogError logs an error event.
func (m *MockLogger) LogError(event string, method string, url string, statusCode int, serverStatusMessage string, err error, rawResponse string) {
	m.Called(event, method, url, statusCode, serverStatusMessage, err, rawResponse)
}

// Example for LogAuthTokenError:
func (m *MockLogger) LogAuthTokenError(event string, method string, url string, statusCode int, err error) {
	m.Called(event, method, url, statusCode, err)
}

// LogCookies logs information about cookies.
func (m *MockLogger) LogCookies(direction string, obj interface{}, method, url string) {
	m.Called(direction, obj, method, url)
}

// LogRetryAttempt logs a retry attempt.
func (m *MockLogger) LogRetryAttempt(event string, method string, url string, attempt int, reason string, waitDuration time.Duration, err error) {
	m.Called(event, method, url, attempt, reason, waitDuration, err)
}

// LogRateLimiting logs rate limiting events.
func (m *MockLogger) LogRateLimiting(event string, method string, url string, retryAfter string, waitDuration time.Duration) {
	m.Called(event, method, url, retryAfter, waitDuration)
}

// LogResponse logs HTTP responses.
func (m *MockLogger) LogResponse(event string, method string, url string, statusCode int, responseBody string, responseHeaders map[string][]string, duration time.Duration) {
	m.Called(event, method, url, statusCode, responseBody, responseHeaders, duration)
}
