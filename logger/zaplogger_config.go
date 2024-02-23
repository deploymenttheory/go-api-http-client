// logger.go
package logger

// Ref: https://betterstack.com/community/guides/logging/go/zap/#logging-errors-with-zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BuildLogger creates and returns a new zap logger instance.
// It configures the logger with JSON formatting and a custom encoder to ensure the 'pid', 'application', and 'timestamp' fields
// appear at the end of each log message. The function panics if the logger cannot be initialized.
func BuildLogger(logLevel LogLevel, encoding string, logConsoleSeparator string) Logger {

	// Set up custom encoder configuration
	encoderCfg := zap.NewProductionEncoderConfig()

	// Time settings
	encoderCfg.TimeKey = "timestamp"                   // Key for enabling serialized time field.
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder // Encodes time in RFC3339 format, which is fully compatible with ISO8601 and more precise.

	// Log level settings
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // Encodes log levels in uppercase with ANSI colors for better visibility in terminal outputs.

	// Caller settings
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder // Encodes the caller in a shortened format `file:line`, making it concise yet informative. / zapcore.FullCallerEncoder for full path

	// Additional settings
	encoderCfg.MessageKey = "msg"                             // Key for the log message.
	encoderCfg.LevelKey = "level"                             // Key for the log level.
	encoderCfg.NameKey = "logger"                             // Key for the logger name.
	encoderCfg.CallerKey = "caller"                           // Key for the caller information.
	encoderCfg.FunctionKey = "func"                           // Key for the function name from where the log was initiated.
	encoderCfg.StacktraceKey = "stacktrace"                   // Key for the stack trace field in case of errors or panics.
	encoderCfg.LineEnding = zapcore.DefaultLineEnding         // Specifies the line ending character(s), defaulting to a newline.
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder // Encodes durations in a human-readable string format.

	// Name and function encoding (optional, depends on logging requirements)
	encoderCfg.EncodeName = zapcore.FullNameEncoder // Encodes the logger's name as-is, without any modifications.

	// Console-specific settings (if using console encoding)
	if encoding == "console" {
		encoderCfg.ConsoleSeparator = logConsoleSeparator
	}

	// Convert the custom LogLevel to zap's logging level
	zapLogLevel := convertToZapLevel(logLevel)

	// Define the logger configuration
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLogLevel), // Default log level is Info
		Development:       false,                             // Set to true if the logger is used in a development environment
		Encoding:          encoding,                          // Supports 'json' and 'console' encodings
		DisableCaller:     true,
		DisableStacktrace: true,
		Sampling:          nil,
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stdout", // Log info and above to standard output
		},
		ErrorOutputPaths: []string{ // is similar to OutputPaths but is used for Zap's internal errors only, not those generated or logged by your application (such as the error from mismatched loosely-typed key/value pairs).
			"stderr", // Log internal Zap errors to standard error
		},
		InitialFields: map[string]interface{}{ // specifies global contextual fields that should be included in every log entry produced by each logger created from the Config object
			//"pid":         os.Getpid(),
			//"application": version.GetAppName(),
		},
	}
	// Build the logger from the configuration
	logger := zap.Must(config.Build())

	// Wrap the original core with the custom core
	wrappedCore := &customCore{logger.Core()}
	wrappedLogger := zap.New(wrappedCore)

	// Wrap the Zap logger in your defaultLogger struct, which implements the Logger interface
	return &defaultLogger{
		logger:   wrappedLogger,
		logLevel: logLevel,
	}
}

// convertToZapLevel converts the custom LogLevel to a zapcore.Level
func convertToZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case LogLevelDebug:
		return zap.DebugLevel
	case LogLevelInfo:
		return zap.InfoLevel
	case LogLevelWarn:
		return zap.WarnLevel
	case LogLevelError:
		return zap.ErrorLevel
	case LogLevelDPanic:
		return zap.DPanicLevel
	case LogLevelPanic:
		return zap.PanicLevel
	case LogLevelFatal:
		return zap.FatalLevel
	default:
		return zap.InfoLevel // Default to InfoLevel
	}
}
