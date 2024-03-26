package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LogRequestStart logs the initiation of an HTTP request if the current log level permits.
func (d *defaultLogger) LogRequestStart(event string, requestID string, userID string, method string, url string, headers map[string][]string) {
	if d.logLevel <= LogLevelInfo {
		fields := []zap.Field{
			zap.String("event", event),
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
func (d *defaultLogger) LogRequestEnd(event string, method string, url string, statusCode int, duration time.Duration) {
	if d.logLevel <= LogLevelInfo {
		fields := []zap.Field{
			zap.String("event", event),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.Duration("duration", duration),
		}
		d.logger.Info("HTTP request completed", fields...)
	}
}

// LogError logs an error that occurs during the processing of an HTTP request or any other event, if the current log level permits.
func (d *defaultLogger) LogError(event string, method, url string, statusCode int, serverStatusMessage string, err error, rawResponse string) {
	if d.logLevel <= LogLevelError {
		errorMessage := ""
		if err != nil {
			errorMessage = err.Error()
		}

		fields := []zap.Field{
			zap.String("event", event),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.String("status_message", serverStatusMessage),
			zap.String("error_message", errorMessage),
			zap.String("raw_response", rawResponse),
		}
		d.logger.Error("Error occurred", fields...)
	}
}

// LogAuthTokenError logs issues encountered during the authentication token acquisition process.
func (d *defaultLogger) LogAuthTokenError(event string, method string, url string, statusCode int, err error) {
	if d.logLevel <= LogLevelError {
		fields := []zap.Field{
			zap.String("event", event),
			zap.String("method", method),
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.String("error_message", err.Error()),
		}
		d.logger.Error("Error obtaining authentication token", fields...)
	}
}

// LogRetryAttempt logs a retry attempt for an HTTP request if the current log level permits, including wait duration and the error that triggered the retry.
func (d *defaultLogger) LogRetryAttempt(event string, method string, url string, attempt int, reason string, waitDuration time.Duration, err error) {
	if d.logLevel <= LogLevelWarn {
		fields := []zap.Field{
			zap.String("event", event),
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
func (d *defaultLogger) LogRateLimiting(event string, method string, url string, retryAfter string, waitDuration time.Duration) {
	if d.logLevel <= LogLevelWarn {
		fields := []zap.Field{
			zap.String("event", event),
			zap.String("method", method),
			zap.String("url", url),
			zap.String("retry_after", retryAfter),
			zap.Duration("wait_duration", waitDuration),
		}
		d.logger.Warn("Rate limit encountered, waiting before retrying", fields...)
	}
}

// LogResponse logs details about an HTTP response if the current log level permits.
func (d *defaultLogger) LogResponse(event string, method string, url string, statusCode int, responseBody string, responseHeaders map[string][]string, duration time.Duration) {
	if d.logLevel <= LogLevelInfo {
		fields := []zap.Field{
			zap.String("event", event),
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

// LogCookies logs the cookies associated with an HTTP request or response.
// `direction` indicates whether the cookies are being sent ("outgoing") or received ("incoming").
// `obj` can be either *http.Request or *http.Response.
func (d *defaultLogger) LogCookies(direction string, obj interface{}, method, url string) {
	var cookies []*http.Cookie
	var objectType string

	// Determine the type and extract cookies
	switch v := obj.(type) {
	case *http.Request:
		cookies = v.Cookies()
		objectType = "request"
	case *http.Response:
		cookies = v.Cookies()
		objectType = "response"
	default:
		// Log a warning if the object is not a request or response
		d.logger.Warn("Invalid object type for cookie logging", zap.Any("object", obj))
		return
	}

	// Log the cookies if any are present
	if len(cookies) > 0 {
		fields := []zap.Field{
			zap.String("direction", direction),
			zap.String("object_type", objectType),
			zap.String("method", method),
			zap.String("url", url),
			zap.Any("cookies", cookies),
		}
		d.logger.Debug("Cookies logged", fields...)
	}
}
