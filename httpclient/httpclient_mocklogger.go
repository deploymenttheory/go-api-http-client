// httpclient_mocklogger.go
package httpclient

import (
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

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

// Error method
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
