// http_error_handling.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/internal/logger"
	"go.uber.org/zap"
)

// APIError represents a structured API error response.
type APIError struct {
	StatusCode int    // HTTP status code
	Type       string // A brief identifier for the type of error (e.g., "RateLimit", "BadRequest", etc.)
	Message    string // Human-readable message
}

// StructuredError represents a structured error response from the API.
type StructuredError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Error returns a string representation of the APIError.
func (e *APIError) Error() string {
	return fmt.Sprintf("API Error (Type: %s, Code: %d): %s", e.Type, e.StatusCode, e.Message)
}

// HandleAPIError handles error responses from the API, converting them into a structured error if possible.
func HandleAPIError(resp *http.Response, log logger.Logger) error {
	var structuredErr StructuredError
	err := json.NewDecoder(resp.Body).Decode(&structuredErr)
	if err == nil && structuredErr.Error.Message != "" {
		// Using structured logging to log the structured error details
		log.Warn("API returned structured error",
			zap.String("status", resp.Status),
			zap.String("error_code", structuredErr.Error.Code),
			zap.String("error_message", structuredErr.Error.Message),
		)
		return &APIError{
			StatusCode: resp.StatusCode,
			Type:       structuredErr.Error.Code,
			Message:    structuredErr.Error.Message,
		}
	}

	var errMsg string
	err = json.NewDecoder(resp.Body).Decode(&errMsg)
	if err != nil || errMsg == "" {
		errMsg = fmt.Sprintf("Unexpected error with status code: %d", resp.StatusCode)
		// Logging with structured fields
		log.Warn("Failed to decode API error message, using default error message",
			zap.String("status", resp.Status),
			zap.String("error_message", errMsg),
		)
	} else {
		// Logging non-structured error as a warning with structured fields
		log.Warn("API returned non-structured error",
			zap.String("status", resp.Status),
			zap.String("error_message", errMsg),
		)
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Type:       "UnexpectedError",
		Message:    errMsg,
	}
}

// TranslateStatusCode provides a human-readable message for HTTP status codes.
func TranslateStatusCode(statusCode int) string {
	messages := map[int]string{
		http.StatusOK:                    "Request successful.",
		http.StatusCreated:               "Request to create or update resource successful.",
		http.StatusAccepted:              "The request was accepted for processing, but the processing has not completed.",
		http.StatusNoContent:             "Request successful. Resource successfully deleted.",
		http.StatusBadRequest:            "Bad request. Verify the syntax of the request.",
		http.StatusUnauthorized:          "Authentication failed. Verify the credentials being used for the request.",
		http.StatusForbidden:             "Invalid permissions. Verify the account being used has the proper permissions for the resource you are trying to access.",
		http.StatusNotFound:              "Resource not found. Verify the URL path is correct.",
		http.StatusConflict:              "Conflict. See the error response for additional details.",
		http.StatusPreconditionFailed:    "Precondition failed. See error description for additional details.",
		http.StatusRequestEntityTooLarge: "Payload too large.",
		http.StatusRequestURITooLong:     "Request-URI too long.",
		http.StatusInternalServerError:   "Internal server error. Retry the request or contact support if the error persists.",
		http.StatusBadGateway:            "Bad Gateway. Generally due to a timeout issue.",
		http.StatusServiceUnavailable:    "Service unavailable.",
	}

	if message, exists := messages[statusCode]; exists {
		return message
	}
	return "An unexpected error occurred. Please try again later."
}

// IsNonRetryableError checks if the provided response indicates a non-retryable error.
func IsNonRetryableError(resp *http.Response) bool {
	// List of non-retryable HTTP status codes
	nonRetryableStatusCodes := map[int]bool{
		http.StatusBadRequest:            true, // 400
		http.StatusUnauthorized:          true, // 401
		http.StatusForbidden:             true, // 403
		http.StatusNotFound:              true, // 404
		http.StatusConflict:              true, // 409
		http.StatusRequestEntityTooLarge: true, // 413
		http.StatusRequestURITooLong:     true, // 414
	}

	_, isNonRetryable := nonRetryableStatusCodes[resp.StatusCode]
	return isNonRetryable
}

// IsRateLimitError checks if the provided response indicates a rate limit error.
func IsRateLimitError(resp *http.Response) bool {
	return resp.StatusCode == http.StatusTooManyRequests
}

// IsTransientError checks if an error or HTTP response indicates a transient error.
func IsTransientError(resp *http.Response) bool {
	transientStatusCodes := map[int]bool{
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
	}
	return resp != nil && transientStatusCodes[resp.StatusCode]
}
