// zaplogger_mocklogger.go
package logger

import (
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockLogger is a mock type for the Logger interface, embedding a *zap.Logger to satisfy the type requirement.
type MockLogger struct {
	mock.Mock
	*zap.Logger
}

// NewMockLogger creates a new instance of MockLogger with an embedded no-op *zap.Logger.
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Logger: zap.NewNop(),
	}
}
