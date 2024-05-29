// zaplogger_structured_messaging.go
package logger

import (
	"go.uber.org/zap"
)

type LogLevel int

const (
	LogLevelDebug  LogLevel = -1
	LogLevelInfo   LogLevel = 0
	LogLevelWarn   LogLevel = 1
	LogLevelError  LogLevel = 2
	LogLevelDPanic LogLevel = 3
	LogLevelPanic  LogLevel = 4
	LogLevelFatal  LogLevel = 5
	LogLevelNone            = 0
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
