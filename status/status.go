// status.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package status

import (
	"net/http"
)

// IsRedirectStatusCode checks if the provided HTTP status code is one of the redirect codes.
// Redirect status codes instruct the client to make a new request to a different URI, as defined in the response's Location header.
//
// - 301 Moved Permanently: The requested resource has been assigned a new permanent URI and any future references to this resource should use one of the returned URIs.
// - 302 Found: The requested resource resides temporarily under a different URI. Since the redirection might be altered on occasion, the client should continue to use the Request-URI for future requests.
// - 303 See Other: The response to the request can be found under a different URI and should be retrieved using a GET method on that resource. This method exists primarily to allow the output of a POST-activated script to redirect the user agent to a selected resource.
// - 307 Temporary Redirect: The requested resource resides temporarily under a different URI. The client should not change the request method if it performs an automatic redirection to that URI.
// - 308 Permanent Redirect: The request and all future requests should be repeated using another URI. 308 parallel the behavior of 301 but do not allow the HTTP method to change. So, for example, submitting a form to a permanently redirected resource may continue smoothly.
//
// The function returns true if the statusCode is one of the above redirect statuses, indicating that the client should follow the redirection as specified in the Location header of the response.
func IsRedirectStatusCode(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

// IsPermanentRedirect checks if the provided HTTP status code is one of the permanent redirect codes.
func IsPermanentRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently,
		http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

// IsNonRetryableStatusCode checks if the provided response indicates a non-retryable error.
func IsNonRetryableStatusCode(resp *http.Response) bool {
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

	_, isNonRetryable := nonRetryableStatusCodes[resp.StatusCode]
	return isNonRetryable
}

// IsTransientError checks if an error or HTTP response indicates a transient error.
func IsTransientError(resp *http.Response) bool {
	transientStatusCodes := map[int]bool{
		http.StatusInternalServerError: true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	}
	return resp != nil && transientStatusCodes[resp.StatusCode]
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

	_, retryable := retryableStatusCodes[statusCode]
	return retryable
}
