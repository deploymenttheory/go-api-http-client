// zaplogger_logfields.go
package logger

import (
	"time"

	"go.uber.org/zap"
)

// LogRequestStart logs the initiation of an HTTP request, including the HTTP method, URL, and headers.
// This function is intended to be called at the beginning of an HTTP request lifecycle.
func LogRequestStart(logger *zap.Logger, requestID string, userID string, method string, url string, headers map[string][]string) {
	fields := []zap.Field{
		zap.String("event", "request_start"),
		zap.String("method", method),
		zap.String("url", url),
		zap.Reflect("headers", headers), // Consider sanitizing or selectively logging headers
	}
	logger.Info("HTTP request started", fields...)
}

// LogRequestEnd logs the completion of an HTTP request, including the HTTP method, URL, status code, and duration.
// This function is intended to be called at the end of an HTTP request lifecycle.
func LogRequestEnd(logger *zap.Logger, method string, url string, statusCode int, duration time.Duration) {
	fields := []zap.Field{
		zap.String("event", "request_end"),
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
	}
	logger.Info("HTTP request completed", fields...)
}

// LogError logs an error that occurs during the processing of an HTTP request, including the HTTP method, URL, status code, error message, and stack trace.
// This function is intended to be called when an error is encountered during an HTTP request lifecycle.
func LogError(logger *zap.Logger, method string, url string, statusCode int, err error, stacktrace string) {
	fields := []zap.Field{
		zap.String("event", "request_error"),
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("status_code", statusCode),
		zap.String("error_message", err.Error()),
		zap.String("stacktrace", stacktrace),
	}
	logger.Error("Error during HTTP request", fields...)
}

// LogRetryAttempt logs a retry attempt for an HTTP request, including the HTTP method, URL, attempt number, and reason for the retry.
// This function is intended to be called when an HTTP request is retried.
func LogRetryAttempt(logger *zap.Logger, method string, url string, attempt int, reason string) {
	fields := []zap.Field{
		zap.String("event", "retry_attempt"),
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("attempt", attempt),
		zap.String("reason", reason),
	}
	logger.Warn("HTTP request retry", fields...)
}

// LogRateLimiting logs when an HTTP request is rate-limited, including the HTTP method, URL, and the value of the 'Retry-After' header.
// This function is intended to be called when an HTTP request encounters rate limiting.
func LogRateLimiting(logger *zap.Logger, method string, url string, retryAfter string) {
	fields := []zap.Field{
		zap.String("event", "rate_limited"),
		zap.String("method", method),
		zap.String("url", url),
		zap.String("retry_after", retryAfter),
	}
	logger.Warn("HTTP request rate-limited", fields...)
}
