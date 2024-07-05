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
	case http.StatusMovedPermanently, // 301
		http.StatusFound,             // 302
		http.StatusSeeOther,          // 303
		http.StatusTemporaryRedirect, // 307
		http.StatusPermanentRedirect: // 308
		return true
	default:
		return false
	}
}

// IsPermanentRedirect checks if the provided HTTP status code is one of the permanent redirect codes.
func IsPermanentRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently, // 301
		http.StatusPermanentRedirect: // 308
		return true
	default:
		return false
	}
}

// IsNonRetryableStatusCode checks if the provided response indicates a non-retryable error.
func IsNonRetryableStatusCode(resp *http.Response) bool {
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
