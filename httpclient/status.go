package httpclient

import "net/http"

// IsNonRetryableStatusCode checks if the provided response indicates a non-retryable error.
func IsNonRetryableStatusCode(statusCode int) bool {
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

	_, isNonRetryable := nonRetryableStatusCodes[statusCode]
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
