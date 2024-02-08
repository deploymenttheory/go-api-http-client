// logger.go
package logger

// Ref: https://betterstack.com/community/guides/logging/go/zap/#logging-errors-with-zap
import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the level of logging. Higher values denote more severe log messages.
type LogLevel int

const (
	// LogLevelDebug is for messages that are useful during software debugging.
	LogLevelDebug LogLevel = iota - 1 // Adjusted value to -1 to follow Zap's convention for DEBUG.
	// LogLevelInfo is for informational messages, indicating normal operation.
	LogLevelInfo
	// LogLevelWarn is for messages that highlight potential issues in the system.
	LogLevelWarn
	// LogLevelError is for messages that highlight errors in the application's execution.
	LogLevelError
	// LogLevelDPanic is for severe error conditions that are actionable in development. It panics in development and logs as an error in production.
	LogLevelDPanic
	// LogLevelPanic is for severe error conditions that should cause the program to panic.
	LogLevelPanic
	// LogLevelFatal is for errors that require immediate program termination.
	LogLevelFatal
	// LogLevelNone is used to disable logging.
	LogLevelNone
)

// Logger interface with structured logging capabilities at various levels.
type Logger interface {
	SetLevel(level LogLevel)
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field) error
	Panic(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	With(fields ...zapcore.Field) Logger
	GetLogLevel() LogLevel
}

// defaultLogger is an implementation of the Logger interface using Uber's zap logging library.
// It provides structured, leveled logging capabilities. The logLevel field controls the verbosity
// of the logs that this logger will produce, allowing filtering of logs based on their importance.
type defaultLogger struct {
	logger   *zap.Logger // logger holds the reference to the zap.Logger instance.
	logLevel LogLevel    // logLevel determines the current logging level (e.g., DEBUG, INFO, WARN).
}

// NewDefaultLogger creates and returns a new logger instance with a default production configuration.
// It uses JSON formatting for log messages and sets the initial log level to Info. If the logger cannot
// be initialized, the function panics to indicate a critical setup failure.
func NewDefaultLogger() Logger {
	// Set up custom encoder configuration
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder // Use ISO8601 format for timestamps

	// Define the logger configuration
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel), // Default log level is Info
		Development:      false,                               // Set to true if the logger is used in a development environment
		Encoding:         "json",                              // Use JSON format for structured logging
		EncoderConfig:    encoderCfg,
		OutputPaths:      []string{"stdout"}, // Log to standard output
		ErrorOutputPaths: []string{"stderr"}, // Log internal Zap errors to standard error
		InitialFields: map[string]interface{}{
			"application": "your-application-name", // Customize this field to suit your needs
		},
	}

	// Build the logger from the configuration
	logger, err := config.Build()
	if err != nil {
		panic(err) // Panic if there is an error initializing the logger
	}

	// Wrap the Zap logger in your defaultLogger struct, which implements your Logger interface
	return &defaultLogger{
		logger:   logger,
		logLevel: LogLevelInfo, // Assuming LogLevelInfo maps to zap.InfoLevel
	}
}

// SetLevel updates the logging level of the logger. It controls the verbosity of the logs,
// allowing the option to filter out less severe messages based on the specified level.
func (d *defaultLogger) SetLevel(level LogLevel) {
	d.logLevel = level
}

// ToZapFields converts a variadic list of key-value pairs into a slice of Zap fields.
// This allows for structured logging with strongly-typed values. The function assumes
// that keys are strings and values can be of any type, leveraging zap.Any for type detection.
func ToZapFields(keysAndValues ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, val := keysAndValues[i], keysAndValues[i+1]
		fields = append(fields, zap.Any(key.(string), val))
	}
	return fields
}

// Debug logs a message at the Debug level. This level is typically used for detailed troubleshooting
// information that is only relevant during active development or debugging.
func (d *defaultLogger) Debug(msg string, fields ...zapcore.Field) {
	if d.logLevel >= LogLevelDebug {
		d.logger.Debug(msg, fields...)
	}
}

// Info logs a message at the Info level. This level is used for informational messages that highlight
// the normal operation of the application.
func (d *defaultLogger) Info(msg string, fields ...zapcore.Field) {
	if d.logLevel >= LogLevelInfo {
		d.logger.Info(msg, fields...)
	}
}

// Warn logs a message at the Warn level. This level is used for potentially harmful situations or to
// indicate that some issues may require attention.
func (d *defaultLogger) Warn(msg string, fields ...zapcore.Field) {
	if d.logLevel >= LogLevelWarn {
		d.logger.Warn(msg, fields...)
	}
}

// Error logs a message at the Error level. This level is used to log error events that might still allow
// the application to continue running.
// Error logs a message at the Error level and returns a formatted error.
func (d *defaultLogger) Error(msg string, fields ...zapcore.Field) error {
	d.logger.Error(msg, fields...)
	return fmt.Errorf(msg)
}

// Panic logs a message at the Panic level and then panics. This level is used to log severe error events
// that will likely lead the application to abort.
func (d *defaultLogger) Panic(msg string, fields ...zapcore.Field) {
	if d.logLevel >= LogLevelPanic {
		d.logger.Panic(msg, fields...)
	}
}

// Fatal logs a message at the Fatal level and then calls os.Exit(1). This level is used to log severe
// error events that will result in the termination of the application.
func (d *defaultLogger) Fatal(msg string, fields ...zapcore.Field) {
	if d.logLevel >= LogLevelFatal {
		d.logger.Fatal(msg, fields...)
	}
}

// With adds contextual key-value pairs to the logger, returning a new logger instance with the context.
// This is useful for creating a logger with common fields that should be included in all subsequent log entries.
func (d *defaultLogger) With(fields ...zapcore.Field) Logger {
	return &defaultLogger{
		logger:   d.logger.With(fields...),
		logLevel: d.logLevel,
	}
}

// GetLogLevel returns the current logging level of the logger. This allows for checking the logger's
// verbosity level programmatically, which can be useful in conditional logging scenarios.
func (d *defaultLogger) GetLogLevel() LogLevel {
	return d.logLevel
}

// GetLoggerBasedOnEnv returns a zap.Logger instance configured for either
// production or development based on the APP_ENV environment variable.
// If APP_ENV is set to "development", it returns a development logger.
// Otherwise, it defaults to a production logger.
func GetLoggerBasedOnEnv() *zap.Logger {
	if os.Getenv("APP_ENV") == "development" {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err) // Handle error according to your application's error policy
		}
		return logger
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err) // Handle error according to your application's error policy
	}
	return logger
}
