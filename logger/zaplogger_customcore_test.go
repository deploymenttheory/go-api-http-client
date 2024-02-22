package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MockCore is a mock implementation of zapcore.Core for testing purposes
type MockCore struct {
	mock.Mock
	zapcore.Core
}

func (m *MockCore) With(fields []zapcore.Field) zapcore.Core {
	args := m.Called(fields)
	return args.Get(0).(zapcore.Core)
}

func (m *MockCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	args := m.Called(entry, fields)
	return args.Error(0)
}

func (m *MockCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	args := m.Called(entry, checkedEntry)
	return args.Get(0).(*zapcore.CheckedEntry)
}

func (m *MockCore) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// TestCustomCoreWith tests the With method of customCore
func TestCustomCoreWith(t *testing.T) {
	mockCore := new(MockCore)
	cCore := &customCore{Core: mockCore}

	// Setup expectations
	mockCore.On("With", mock.Anything).Return(cCore)

	fields := []zapcore.Field{zap.String("key", "value")}
	newCore := cCore.With(fields)

	// Assertions
	mockCore.AssertCalled(t, "With", fields)
	assert.IsType(t, &customCore{}, newCore, "Expected newCore to be of type *customCore")
}

// TestCustomCoreWrite tests the Write method, particularly the reordering logic
func TestCustomCoreWrite(t *testing.T) {
	mockCore := new(MockCore)
	cCore := &customCore{Core: mockCore}

	// Setup expectations
	mockCore.On("Write", mock.Anything, mock.AnythingOfType("[]zapcore.Field")).Return(nil)

	entry := zapcore.Entry{}
	fields := []zapcore.Field{
		zap.String("key", "value"),
		zap.Int("pid", 1234),
		zap.String("application", "testApp"),
		zap.String("anotherKey", "anotherValue"),
	}

	err := cCore.Write(entry, fields)

	// Assertions
	assert.NoError(t, err)
	mockCore.AssertCalled(t, "Write", entry, mock.MatchedBy(func(f []zapcore.Field) bool {
		// Verify pid and application are at the end
		return len(f) >= 2 && f[len(f)-2].Key == "pid" && f[len(f)-1].Key == "application"
	}))
}

// TestCustomCoreSync tests the Sync method of customCore
func TestCustomCoreSync(t *testing.T) {
	mockCore := new(MockCore)
	cCore := &customCore{Core: mockCore}

	// Setup expectations
	mockCore.On("Sync").Return(nil)

	err := cCore.Sync()

	// Assertions
	assert.NoError(t, err)
	mockCore.AssertCalled(t, "Sync")
}
