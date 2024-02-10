// logger.go
package logger

// Ref: https://betterstack.com/community/guides/logging/go/zap/#logging-errors-with-zap

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogOutputJSON          = "json"
	LogOutputHumanReadable = "human-readable"
)

// BuildLogger creates and returns a new zap logger instance.
// It configures the logger with JSON formatting and a custom encoder to ensure the 'pid', 'application', and 'timestamp' fields
// appear at the end of each log message. The function panics if the logger cannot be initialized.
func BuildJSONLogger(logLevel LogLevel, logOutputFormat string) Logger {

	// Set up custom encoder configuration
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder // Use ISO8601 format for timestamps

	// Convert the custom LogLevel to zap's logging level
	zapLogLevel := convertToZapLevel(logLevel)

	// Define the logger configuration
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLogLevel), // Default log level is Info
		Development:       false,                             // Set to true if the logger is used in a development environment
		Encoding:          "console",                         // Supports 'json' and 'console' encodings
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

func NewHumanReadableLogger(logLevel LogLevel, logOutputFormat string) Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if logOutputFormat == LogOutputHumanReadable {
		// For human-readable format, use a console encoder with customized level encoder
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		// Default to JSON encoder for structured logging
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), convertToZapLevel(logLevel))

	// Wrap the core with customCore to maintain field reordering logic
	wrappedCore := &customCore{Core: core}

	logger := zap.New(wrappedCore, zap.AddCaller())

	return &defaultLogger{
		logger:   logger,
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
