package httpclient

import (
	"go.uber.org/zap"
)

type LogLevel int

const (
	LogLevelNone LogLevel = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarning
	LogLevelError
	LogLevelPanic
	LogLevelFatal
)

// Logger interface as defined earlier
type Logger interface {
	SetLevel(level LogLevel)
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Panic(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
}

// defaultLogger is an implementation of the Logger interface using Uber's zap logging library.
// It provides structured, leveled logging capabilities. The logLevel field controls the verbosity
// of the logs that this logger will produce, allowing filtering of logs based on their importance.
type defaultLogger struct {
	logger   *zap.Logger // logger holds the reference to the zap.Logger instance.
	logLevel LogLevel    // logLevel determines the current logging level (e.g., DEBUG, INFO, WARN).
}

// NewDefaultLogger initializes and returns a new instance of defaultLogger with a production
// configuration from the zap logging library. This function sets the default logging level to
// LogLevelWarning, which means that by default, DEBUG and INFO logs will be suppressed.
// In case of an error while initializing the zap.Logger, this function will panic, as the
// inability to log is considered a fatal error in production environments.
func NewDefaultLogger() Logger {
	logger, err := zap.NewProduction() // Initialize a zap logger with production settings.
	if err != nil {
		panic(err) // Panic if there is an error initializing the logger, as logging is critical.
	}

	return &defaultLogger{
		logger:   logger,          // Set the initialized zap.Logger.
		logLevel: LogLevelWarning, // Set the default log level to warning.
	}
}

// Implement the SetLevel method for defaultLogger
func (d *defaultLogger) SetLevel(level LogLevel) {
	d.logLevel = level
}

// Convert keysAndValues to zap.Fields
func toZapFields(keysAndValues ...interface{}) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, val := keysAndValues[i], keysAndValues[i+1]
		fields = append(fields, zap.Any(key.(string), val))
	}
	return fields
}

// Debug method implementation
func (d *defaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	if d.logLevel >= LogLevelDebug {
		d.logger.Debug(msg, toZapFields(keysAndValues...)...)
	}
}

// Info method implementation
func (d *defaultLogger) Info(msg string, keysAndValues ...interface{}) {
	if d.logLevel >= LogLevelInfo {
		d.logger.Info(msg, toZapFields(keysAndValues...)...)
	}
}

// Warn method implementation
func (d *defaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	if d.logLevel >= LogLevelWarning {
		d.logger.Warn(msg, toZapFields(keysAndValues...)...)
	}
}

// Error method implementation
func (d *defaultLogger) Error(msg string, keysAndValues ...interface{}) {
	if d.logLevel > LogLevelNone {
		d.logger.Error(msg, toZapFields(keysAndValues...)...)
	}
}

// Panic method implementation
func (d *defaultLogger) Panic(msg string, keysAndValues ...interface{}) {
	if d.logLevel >= LogLevelPanic {
		d.logger.Panic(msg, toZapFields(keysAndValues...)...)
	}
}

// Fatal method implementation
func (d *defaultLogger) Fatal(msg string, keysAndValues ...interface{}) {
	if d.logLevel >= LogLevelFatal {
		d.logger.Fatal(msg, toZapFields(keysAndValues...)...)
	}
}
