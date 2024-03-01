// zaplogger_logger.go
// Ref: https://betterstack.com/community/guides/logging/go/zap/#logging-errors-with-zap
package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// defaultLogger is an implementation of the Logger interface using Uber's zap logging library.
// It provides structured, leveled logging capabilities. The logLevel field controls the verbosity
// of the logs that this logger will produce, allowing filtering of logs based on their importance.
type defaultLogger struct {
	logger   *zap.Logger // logger holds the reference to the zap.Logger instance.
	logLevel LogLevel    // logLevel determines the current logging level (e.g., DEBUG, INFO, WARN).
}

// Logger interface with structured logging capabilities at various levels.
type Logger interface {
	GetLogLevel() LogLevel
	SetLevel(level LogLevel)
	With(fields ...zapcore.Field) Logger
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field) error
	Panic(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)

	// Updated method signatures to include the 'event' parameter
	LogRequestStart(event string, requestID string, userID string, method string, url string, headers map[string][]string)
	LogRequestEnd(event string, method string, url string, statusCode int, duration time.Duration)
	LogError(event string, method string, url string, statusCode int, serverStatusMessage string, err error, rawResponse string)
	LogAuthTokenError(event string, method string, url string, statusCode int, err error)
	LogRetryAttempt(event string, method string, url string, attempt int, reason string, waitDuration time.Duration, err error)
	LogRateLimiting(event string, method string, url string, retryAfter string, waitDuration time.Duration)
	LogResponse(event string, method string, url string, statusCode int, responseBody string, responseHeaders map[string][]string, duration time.Duration)
}

// GetLogLevel returns the current logging level of the logger. This allows for checking the logger's
// verbosity level programmatically, which can be useful in conditional logging scenarios.
func (d *defaultLogger) GetLogLevel() LogLevel {
	return d.logLevel
}

// SetLevel updates the logging level of the logger. It controls the verbosity of the logs,
// allowing the option to filter out less severe messages based on the specified level.
func (d *defaultLogger) SetLevel(level LogLevel) {
	d.logLevel = level
}

// With adds contextual key-value pairs to the logger, returning a new logger instance with the context.
// This is useful for creating a logger with common fields that should be included in all subsequent log entries.
func (d *defaultLogger) With(fields ...zapcore.Field) Logger {
	return &defaultLogger{
		logger:   d.logger.With(fields...),
		logLevel: d.logLevel,
	}
}

// Debug logs a message at the Debug level. This level is typically used for detailed troubleshooting
// information that is only relevant during active development or debugging.
func (d *defaultLogger) Debug(msg string, fields ...zapcore.Field) {
	if d.logLevel <= LogLevelDebug {
		d.logger.Debug(msg, fields...)
	}
}

// Info logs a message at the Info level. This level is used for informational messages that highlight
// the normal operation of the application.
func (d *defaultLogger) Info(msg string, fields ...zapcore.Field) {
	if d.logLevel <= LogLevelInfo {
		d.logger.Info(msg, fields...)
	}
}

// Warn logs a message at the Warn level. This level is used for potentially harmful situations or to
// indicate that some issues may require attention.
func (d *defaultLogger) Warn(msg string, fields ...zapcore.Field) {
	if d.logLevel <= LogLevelWarn {
		d.logger.Warn(msg, fields...)
	}
}

// Error logs a message at the Error level. This level is used to log error events that might still allow
// the application to continue running.
// Error logs a message at the Error level and returns a formatted error.
func (d *defaultLogger) Error(msg string, fields ...zapcore.Field) error {
	if d.logLevel <= LogLevelError {
		d.logger.Error(msg, fields...)
	}
	return fmt.Errorf(msg)
}

// Panic logs a message at the Panic level and then panics. This level is used to log severe error events
// that will likely lead the application to abort.
func (d *defaultLogger) Panic(msg string, fields ...zapcore.Field) {
	if d.logLevel <= LogLevelPanic {
		d.logger.Panic(msg, fields...)
	}
}

// Fatal logs a message at the Fatal level and then calls os.Exit(1). This level is used to log severe
// error events that will result in the termination of the application.
func (d *defaultLogger) Fatal(msg string, fields ...zapcore.Field) {
	if d.logLevel <= LogLevelFatal {
		d.logger.Fatal(msg, fields...)
	}
}

// GetLoggerBasedOnEnv returns a zap.Logger instance configured for either
// production or development based on the APP_ENV environment variable.
// If APP_ENV is set to "development", it returns a development logger.
// Otherwise, it defaults to a production logger.
func GetLoggerBasedOnEnv() *zap.Logger {
	if os.Getenv("APP_ENV") == "development" {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		return logger
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return logger
}
