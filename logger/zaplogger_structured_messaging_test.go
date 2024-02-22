// zaplogger_structured_messaging_test.go
package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

// TestToZapFields verifies the functionality of the ToZapFields function, ensuring it correctly converts
// a variadic list of key-value pairs into a slice of Zap fields for structured logging. This test checks
// that the function handles various value types and correctly associates keys with their corresponding values,
// thereby enabling structured logging with strongly-typed values. It uses both string and integer types to
// validate type handling and asserts that the resulting slice of fields matches the expected structure.
func TestToZapFields(t *testing.T) {
	key1 := "key1"
	value1 := "value1"
	key2 := "key2"
	value2 := 123 // Int value to test type handling

	fields := ToZapFields(key1, value1, key2, value2)

	// Verify the length of the resulting fields slice to ensure all key-value pairs were processed
	assert.Len(t, fields, 2, "Expected number of fields does not match")

	// Assert each field matches the expected key-value pair, using zap.String for strings and zap.Any for other types
	assert.Equal(t, zap.String(key1, value1), fields[0], "First field does not match expected key-value pair")
	assert.Equal(t, zap.Any(key2, value2), fields[1], "Second field does not match expected key-value pair")
}
