// zaplogger_structured_messaging.go
package logger

import (
	"go.uber.org/zap"
)

// LogLevel represents the level of logging. Higher values denote more severe log messages.
type LogLevel int

const (
	// LogLevelDebug is for messages that are useful during software debugging.
	LogLevelDebug LogLevel = -1 // Zap's DEBUG level

	// LogLevelInfo is for informational messages, indicating normal operation.
	LogLevelInfo LogLevel = 0 // Zap's INFO level

	// LogLevelWarn is for messages that highlight potential issues in the system.
	LogLevelWarn LogLevel = 1 // Zap's WARN level

	// LogLevelError is for messages that highlight errors in the application's execution.
	LogLevelError LogLevel = 2 // Zap's ERROR level

	// LogLevelDPanic is for severe error conditions that are actionable in development.
	LogLevelDPanic LogLevel = 3 // Zap's DPANIC level

	// LogLevelPanic is for severe error conditions that should cause the program to panic.
	LogLevelPanic LogLevel = 4 // Zap's PANIC level

	// LogLevelFatal is for errors that require immediate program termination.
	LogLevelFatal LogLevel = 5 // Zap's FATAL level

	LogLevelNone = 0
)

// ParseLogLevelFromString takes a string representation of the log level and returns the corresponding LogLevel.
// Used to convert a string log level from a configuration file to a strongly-typed LogLevel.
func ParseLogLevelFromString(levelStr string) LogLevel {
	switch levelStr {
	case "LogLevelDebug":
		return LogLevelDebug
	case "LogLevelInfo":
		return LogLevelInfo
	case "LogLevelWarn":
		return LogLevelWarn
	case "LogLevelError":
		return LogLevelError
	case "LogLevelDPanic":
		return LogLevelDPanic
	case "LogLevelPanic":
		return LogLevelPanic
	case "LogLevelFatal":
		return LogLevelFatal
	default:
		return LogLevelNone
	}
}

// ToZapFields converts a variadic list of key-value pairs into a slice of Zap fields.
// This allows for structured logging with strongly-typed values. The function assumes
// that keys are strings and values can be of any type, leveraging zap.Any for type detection.

// QUERY What does this do? Why is it in steps of two? Surely we can ammend this to accept a [{k, v}, {k, v}] approach?
func ToZapFields(keysAndValues ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, val := keysAndValues[i], keysAndValues[i+1]
		fields = append(fields, zap.Any(key.(string), val))
	}
	return fields
}
