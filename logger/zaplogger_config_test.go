// logger.go
package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestConvertToZapLevel tests the conversion from custom LogLevel to zapcore.Level
func TestConvertToZapLevel(t *testing.T) {
	tests := []struct {
		name          string
		inputLevel    LogLevel
		expectedLevel zapcore.Level
	}{
		{"DebugLevel", LogLevelDebug, zap.DebugLevel},
		{"InfoLevel", LogLevelInfo, zap.InfoLevel},
		{"WarnLevel", LogLevelWarn, zap.WarnLevel},
		{"ErrorLevel", LogLevelError, zap.ErrorLevel},
		{"DPanicLevel", LogLevelDPanic, zap.DPanicLevel},
		{"PanicLevel", LogLevelPanic, zap.PanicLevel},
		{"FatalLevel", LogLevelFatal, zap.FatalLevel},
		{"UnknownLevel", LogLevel(999), zap.InfoLevel}, // Testing default case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToZapLevel(tt.inputLevel)
			assert.Equal(t, tt.expectedLevel, result)
		})
	}
}
