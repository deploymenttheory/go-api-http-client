// httpclient_mocklogger.go
package httpclient

import (
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockLogger is a mock type for the logger.Logger interface used in the httpclient package.
type MockLogger struct {
	mock.Mock
	*zap.Logger
}

// Ensure MockLogger implements the logger.Logger interface.
var _ logger.Logger = (*MockLogger)(nil)

// NewMockLogger creates a new instance of MockLogger with an embedded no-op *zap.Logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Logger: zap.NewNop(),
	}
}

// Define all methods from the logger.Logger interface with mock implementations.
func (m *MockLogger) SetLevel(level logger.LogLevel) {
	m.Called(level)
}

// Mock implementations for unstructured logging methods

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) error {
	args := m.Called(msg, fields)
	return args.Error(0)
}

func (m *MockLogger) Panic(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zap.Field) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) GetLogLevel() logger.LogLevel {
	args := m.Called()
	return args.Get(0).(logger.LogLevel)
}

// Mock implementations for structured logging methods

func (m *MockLogger) LogRequestStart(event string, requestID string, userID string, method string, url string, headers map[string][]string) {
	m.Called(event, requestID, userID, method, url, headers)
}

func (m *MockLogger) LogRequestEnd(event string, method string, url string, statusCode int, duration time.Duration) {
	m.Called(event, method, url, statusCode, duration)
}

func (m *MockLogger) LogError(event string, method string, url string, statusCode int, serverStatusMessage string, err error, rawResponse string) {
	m.Called(event, method, url, statusCode, serverStatusMessage, err, rawResponse)
}

func (m *MockLogger) LogAuthTokenError(event string, method string, url string, statusCode int, err error) {
	m.Called(event, method, url, statusCode, err)
}

func (m *MockLogger) LogRetryAttempt(event string, method string, url string, attempt int, reason string, waitDuration time.Duration, err error) {
	m.Called(event, method, url, attempt, reason, waitDuration, err)
}

func (m *MockLogger) LogRateLimiting(event string, method string, url string, retryAfter string, waitDuration time.Duration) {
	m.Called(event, method, url, retryAfter, waitDuration)
}

func (m *MockLogger) LogResponse(event string, method string, url string, statusCode int, responseBody string, responseHeaders map[string][]string, duration time.Duration) {
	m.Called(event, method, url, statusCode, responseBody, responseHeaders, duration)
}
