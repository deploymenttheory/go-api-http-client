// http_error_handling.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/logger"
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
func HandleAPIError(resp *http.Response, log logger.Logger) *APIError {
	var structuredErr StructuredError
	err := json.NewDecoder(resp.Body).Decode(&structuredErr)
	if err == nil && structuredErr.Error.Message != "" {
		// Log the structured error details
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

	// Default error message for non-structured responses or decode failures
	errMsg := fmt.Sprintf("Unexpected error with status code: %d", resp.StatusCode)
	log.Error("Failed to decode API error message, using default error message",
		zap.String("status", resp.Status),
		zap.String("error_message", errMsg),
	)

	return &APIError{
		StatusCode: resp.StatusCode,
		Type:       "UnexpectedError",
		Message:    errMsg,
	}
}

// TranslateStatusCode provides a human-readable message for HTTP status codes.
func TranslateStatusCode(resp *http.Response) string {

	if resp == nil {
		return "No status code received, possible network or connection error."
	}

	messages := map[int]string{
		http.StatusOK:                            "Request successful.",
		http.StatusCreated:                       "Request to create or update resource successful.",
		http.StatusAccepted:                      "The request was accepted for processing, but the processing has not completed.",
		http.StatusNoContent:                     "Request successful. No content to send for this request.",
		http.StatusBadRequest:                    "Bad request. Verify the syntax of the request.",
		http.StatusUnauthorized:                  "Authentication failed. Verify the credentials being used for the request.",
		http.StatusPaymentRequired:               "Payment required. Access to the requested resource requires payment.",
		http.StatusForbidden:                     "Invalid permissions. Verify the account has the proper permissions for the resource.",
		http.StatusNotFound:                      "Resource not found. Verify the URL path is correct.",
		http.StatusMethodNotAllowed:              "Method not allowed. The method specified is not allowed for the resource.",
		http.StatusNotAcceptable:                 "Not acceptable. The server cannot produce a response matching the list of acceptable values.",
		http.StatusProxyAuthRequired:             "Proxy authentication required. You must authenticate with a proxy server before this request can be served.",
		http.StatusRequestTimeout:                "Request timeout. The server timed out waiting for the request.",
		http.StatusConflict:                      "Conflict. The request could not be processed because of conflict in the request.",
		http.StatusGone:                          "Gone. The resource requested is no longer available and will not be available again.",
		http.StatusLengthRequired:                "Length required. The request did not specify the length of its content, which is required by the requested resource.",
		http.StatusPreconditionFailed:            "Precondition failed. The server does not meet one of the preconditions specified in the request.",
		http.StatusRequestEntityTooLarge:         "Payload too large. The request is larger than the server is willing or able to process.",
		http.StatusRequestURITooLong:             "Request-URI too long. The URI provided was too long for the server to process.",
		http.StatusUnsupportedMediaType:          "Unsupported media type. The request entity has a media type which the server or resource does not support.",
		http.StatusRequestedRangeNotSatisfiable:  "Requested range not satisfiable. The client has asked for a portion of the file, but the server cannot supply that portion.",
		http.StatusExpectationFailed:             "Expectation failed. The server cannot meet the requirements of the Expect request-header field.",
		http.StatusUnprocessableEntity:           "Unprocessable entity. The server understands the content type and syntax of the request but was unable to process the contained instructions.",
		http.StatusLocked:                        "Locked. The resource that is being accessed is locked.",
		http.StatusFailedDependency:              "Failed dependency. The request failed because it depended on another request and that request failed.",
		http.StatusUpgradeRequired:               "Upgrade required. The client should switch to a different protocol.",
		http.StatusPreconditionRequired:          "Precondition required. The server requires that the request be conditional.",
		http.StatusTooManyRequests:               "Too many requests. The user has sent too many requests in a given amount of time.",
		http.StatusRequestHeaderFieldsTooLarge:   "Request header fields too large. The server is unwilling to process the request because its header fields are too large.",
		http.StatusUnavailableForLegalReasons:    "Unavailable for legal reasons. The server is denying access to the resource as a consequence of a legal demand.",
		http.StatusInternalServerError:           "Internal server error. The server encountered an unexpected condition that prevented it from fulfilling the request.",
		http.StatusNotImplemented:                "Not implemented. The server does not support the functionality required to fulfill the request.",
		http.StatusBadGateway:                    "Bad gateway. The server received an invalid response from the upstream server while trying to fulfill the request.",
		http.StatusServiceUnavailable:            "Service unavailable. The server is currently unable to handle the request due to temporary overloading or maintenance.",
		http.StatusGatewayTimeout:                "Gateway timeout. The server did not receive a timely response from the upstream server.",
		http.StatusHTTPVersionNotSupported:       "HTTP version not supported. The server does not support the HTTP protocol version used in the request.",
		http.StatusNetworkAuthenticationRequired: "Network authentication required. The client needs to authenticate to gain network access.",
	}

	// Lookup and return the message for the given status code
	if message, exists := messages[resp.StatusCode]; exists {
		return message
	}
	return fmt.Sprintf("Unknown status code: %d", resp.StatusCode)
}

// IsNonRetryableStatusCode checks if the provided response indicates a non-retryable error.
func IsNonRetryableStatusCode(resp *http.Response) bool {
	// Expanded list of non-retryable HTTP status codes
	nonRetryableStatusCodes := map[int]bool{
		http.StatusBadRequest:                   true, // 400 - Bad Request
		http.StatusUnauthorized:                 true, // 401 - Unauthorized
		http.StatusPaymentRequired:              true, // 402 - Payment Required
		http.StatusForbidden:                    true, // 403 - Forbidden
		http.StatusNotFound:                     true, // 404 - Not Found
		http.StatusMethodNotAllowed:             true, // 405 - Method Not Allowed
		http.StatusNotAcceptable:                true, // 406 - Not Acceptable
		http.StatusProxyAuthRequired:            true, // 407 - Proxy Authentication Required
		http.StatusConflict:                     true, // 409 - Conflict
		http.StatusGone:                         true, // 410 - Gone
		http.StatusLengthRequired:               true, // 411 - Length Required
		http.StatusPreconditionFailed:           true, // 412 - Precondition Failed
		http.StatusRequestEntityTooLarge:        true, // 413 - Request Entity Too Large
		http.StatusRequestURITooLong:            true, // 414 - Request-URI Too Long
		http.StatusUnsupportedMediaType:         true, // 415 - Unsupported Media Type
		http.StatusRequestedRangeNotSatisfiable: true, // 416 - Requested Range Not Satisfiable
		http.StatusExpectationFailed:            true, // 417 - Expectation Failed
		http.StatusUnprocessableEntity:          true, // 422 - Unprocessable Entity
		http.StatusLocked:                       true, // 423 - Locked
		http.StatusFailedDependency:             true, // 424 - Failed Dependency
		http.StatusUpgradeRequired:              true, // 426 - Upgrade Required
		http.StatusPreconditionRequired:         true, // 428 - Precondition Required
		http.StatusRequestHeaderFieldsTooLarge:  true, // 431 - Request Header Fields Too Large
		http.StatusUnavailableForLegalReasons:   true, // 451 - Unavailable For Legal Reasons
	}

	_, isNonRetryable := nonRetryableStatusCodes[resp.StatusCode]
	return isNonRetryable
}

// IsRateLimitError checks if the provided response indicates a rate limit error.
func IsRateLimitError(resp *http.Response) bool {
	if resp == nil {
		// If the response is nil, it cannot be a rate limit error.
		return false
	}
	return resp.StatusCode == http.StatusTooManyRequests
}

// IsTransientError checks if an error or HTTP response indicates a transient error.
func IsTransientError(resp *http.Response) bool {
	transientStatusCodes := map[int]bool{
		http.StatusInternalServerError: true, // 500 Internal Server Error
		http.StatusBadGateway:          true, // 502 Bad Gateway
		http.StatusServiceUnavailable:  true, // 503 Service Unavailable
		http.StatusGatewayTimeout:      true, // 504 - Gateway Timeout
	}
	return resp != nil && transientStatusCodes[resp.StatusCode]
}

// IsRetryableStatusCode checks if the provided HTTP status code is considered retryable.
func IsRetryableStatusCode(statusCode int) bool {
	retryableStatusCodes := map[int]bool{
		http.StatusRequestTimeout:      true, // 408 - Request Timeout
		http.StatusTooManyRequests:     true, // 429
		http.StatusInternalServerError: true, // 500
		http.StatusBadGateway:          true, // 502
		http.StatusServiceUnavailable:  true, // 503
		http.StatusGatewayTimeout:      true, // 504
	}

	_, retryable := retryableStatusCodes[statusCode]
	return retryable
}
