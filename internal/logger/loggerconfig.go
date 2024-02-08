// logger.go
package logger

// Ref: https://betterstack.com/community/guides/logging/go/zap/#logging-errors-with-zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BuildLogger creates and returns a new logger instance with a default production configuration.
// It uses JSON formatting for log messages and sets the initial log level to Info. If the logger cannot
// be initialized, the function panics to indicate a critical setup failure.
func BuildLogger(logLevel LogLevel) Logger {
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
		Encoding:          "json",                            // Use JSON format for structured logging
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"stdout", // Log info and above to standard output
		},
		ErrorOutputPaths: []string{
			"stderr", // Log internal Zap errors to standard error
		},
		InitialFields: map[string]interface{}{
			"application": "your-application-name", // Customize this field to suit your needs
		},
	}

	// Build the logger from the configuration
	logger := zap.Must(config.Build())

	// Wrap the Zap logger in your defaultLogger struct, which implements your Logger interface
	return &defaultLogger{
		logger:   logger,
		logLevel: LogLevelInfo, // Assuming LogLevelInfo maps to zap.InfoLevel
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
