// status.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package response

import (
	"net/http"
)

// IsRedirectStatusCode checks if the provided HTTP status code is one of the redirect codes.
func IsRedirectStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	}

	return false
}

// IsPermanentRedirect checks if the provided HTTP status code is one of the permanent redirect codes.
func IsPermanentRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently,
		http.StatusPermanentRedirect:
		return true
	}

	return false
}

// IsNonRetryableStatusCode checks if the provided response indicates a non-retryable error.
func IsNonRetryableStatusCode(statusCode int) bool {
	nonRetryableStatusCodes := map[int]bool{
		http.StatusBadRequest:                   true,
		http.StatusUnauthorized:                 true,
		http.StatusPaymentRequired:              true,
		http.StatusForbidden:                    true,
		http.StatusNotFound:                     true,
		http.StatusMethodNotAllowed:             true,
		http.StatusNotAcceptable:                true,
		http.StatusProxyAuthRequired:            true,
		http.StatusConflict:                     true,
		http.StatusGone:                         true,
		http.StatusLengthRequired:               true,
		http.StatusPreconditionFailed:           true,
		http.StatusRequestEntityTooLarge:        true,
		http.StatusRequestURITooLong:            true,
		http.StatusUnsupportedMediaType:         true,
		http.StatusRequestedRangeNotSatisfiable: true,
		http.StatusExpectationFailed:            true,
		http.StatusUnprocessableEntity:          true,
		http.StatusLocked:                       true,
		http.StatusFailedDependency:             true,
		http.StatusUpgradeRequired:              true,
		http.StatusPreconditionRequired:         true,
		http.StatusRequestHeaderFieldsTooLarge:  true,
		http.StatusUnavailableForLegalReasons:   true,
	}
	return nonRetryableStatusCodes[statusCode]
}

// IsTransientError checks if an error or HTTP response indicates a transient error.
func IsTransientError(statusCode int) bool {
	transientStatusCodes := map[int]bool{
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}
	return transientStatusCodes[statusCode]
}

// IsRetryableStatusCode checks if the provided HTTP status code is considered retryable.
func IsRetryableStatusCode(statusCode int) bool {
	retryableStatusCodes := map[int]bool{
		http.StatusRequestTimeout:      true,
		http.StatusTooManyRequests:     true,
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}

	return retryableStatusCodes[statusCode]
}
