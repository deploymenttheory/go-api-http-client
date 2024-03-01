package logger

import (
	"time"

	"go.uber.org/zap"
)

// LogRequestStart logs the initiation of an HTTP request if the current log level permits.
func (d *defaultLogger) LogRequestStart(requestID string, userID string, method string, url string, headers map[string][]string) {
	if d.logLevel <= LogLevelInfo {
		fields := []zap.Field{
			zap.String("event", "request_start"),
			zap.String("method", method),
			zap.String("url", url),
			zap.String("request_id", requestID),
			zap.String("user_id", userID),
			zap.Reflect("headers", headers),
		}
		d.logger.Info("HTTP request started", fields...)
	}
}

// LogRequestEnd logs the completion of an HTTP request if the current log level permits.
func (d *defaultLogger) LogRequestEnd(method string, url string, statusCode int, duration time.Duration) {
	if d.logLevel <= LogLevelInfo {
		fields := []zap.Field{
			zap.String("event", "request_end"),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.Duration("duration", duration),
		}
		d.logger.Info("HTTP request completed", fields...)
	}
}

// LogError logs an error that occurs during the processing of an HTTP request if the current log level permits.
func (d *defaultLogger) LogError(method string, url string, statusCode int, err error, stacktrace string) {
	if d.logLevel <= LogLevelError {
		fields := []zap.Field{
			zap.String("event", "request_error"),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.String("error_message", err.Error()),
			zap.String("stacktrace", stacktrace),
		}
		d.logger.Error("Error during HTTP request", fields...)
	}
}

// LogRetryAttempt logs a retry attempt for an HTTP request if the current log level permits, including wait duration and the error that triggered the retry.
func (d *defaultLogger) LogRetryAttempt(method string, url string, attempt int, reason string, waitDuration time.Duration, err error) {
	if d.logLevel <= LogLevelWarn {
		fields := []zap.Field{
			zap.String("event", "retry_attempt"),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("attempt", attempt),
			zap.String("reason", reason),
			zap.Duration("waitDuration", waitDuration),
			zap.String("error_message", err.Error()),
		}
		d.logger.Warn("HTTP request retry", fields...)
	}
}

// LogRateLimiting logs when an HTTP request is rate-limited, including the HTTP method, URL, the value of the 'Retry-After' header, and the actual wait duration.
func (d *defaultLogger) LogRateLimiting(method string, url string, retryAfter string, waitDuration time.Duration) {
	if d.logLevel <= LogLevelWarn {
		fields := []zap.Field{
			zap.String("event", "rate_limited"),
			zap.String("method", method),
			zap.String("url", url),
			zap.String("retry_after", retryAfter),
			zap.Duration("wait_duration", waitDuration),
		}
		d.logger.Warn("Rate limit encountered, waiting before retrying", fields...)
	}
}

// LogResponse logs details about an HTTP response if the current log level permits.
func (d *defaultLogger) LogResponse(method string, url string, statusCode int, responseBody string, responseHeaders map[string][]string, duration time.Duration) {
	if d.logLevel <= LogLevelInfo {
		fields := []zap.Field{
			zap.String("event", "response_received"),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.String("response_body", responseBody),
			zap.Reflect("response_headers", responseHeaders),
			zap.Duration("duration", duration),
		}
		d.logger.Info("HTTP response details", fields...)
	}
}
