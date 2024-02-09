package logger

import (
	"go.uber.org/zap/zapcore"
)

type customCore struct {
	zapcore.Core
}

// With adds structured context to the Core. This method can be used to add additional context or to reorder fields as needed.
func (c *customCore) With(fields []zapcore.Field) zapcore.Core {
	// For simplicity, we're just passing it through in this example
	return &customCore{c.Core.With(fields)}
}

// Write serializes the Entry and any Fields supplied at the log site and writes them to their destination.
func (c *customCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Here you can reorder fields or add custom logic before logging
	// For example, ensure pid and application fields are moved to the end

	// Find and move pid and application to the end
	var pidField, appField zapcore.Field
	var otherFields []zapcore.Field
	for _, field := range fields {
		if field.Key == "pid" {
			pidField = field
		} else if field.Key == "application" {
			appField = field
		} else {
			otherFields = append(otherFields, field)
		}
	}
	reorderedFields := append(otherFields, pidField, appField) // Reorder fields

	return c.Core.Write(entry, reorderedFields)
}

// Check determines whether the supplied Entry should be logged.
func (c *customCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return c.Core.Check(entry, checkedEntry)
}

// Sync flushes buffered logs (if any).
func (c *customCore) Sync() error {
	return c.Core.Sync()
}
