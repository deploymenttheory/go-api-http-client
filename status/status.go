// status.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package status

import (
	"fmt"
	"net/http"
)

// TranslateStatusCode provides a human-readable message for HTTP status codes.
func TranslateStatusCode(resp *http.Response) string {

	if resp == nil {
		return "No status code received, possible network or connection error."
	}

	messages := map[int]string{
		// Successful responses (200-299)
		http.StatusOK:        "Request successful.",
		http.StatusCreated:   "Request to create or update resource successful.",
		http.StatusAccepted:  "The request was accepted for processing, but the processing has not completed.",
		http.StatusNoContent: "Request successful. No content to send for this request.",

		// Redirect status codes (300-399)
		http.StatusMovedPermanently:  "Moved Permanently. The requested resource has been assigned a new permanent URI. Future references should use the returned URI.",
		http.StatusFound:             "Found. The requested resource resides temporarily under a different URI. The client should use the Request-URI for future requests.",
		http.StatusSeeOther:          "See Other. The response to the request can be found under a different URI. A GET method should be used to retrieve the resource.",
		http.StatusTemporaryRedirect: "Temporary Redirect. The requested resource resides temporarily under a different URI. The request method should not change.",
		http.StatusPermanentRedirect: "Permanent Redirect. The requested resource has been permanently moved to a new URI. The request method should not change.",

		// Client error responses (400-499)
		http.StatusBadRequest:                   "Bad request. Verify the syntax of the request.",
		http.StatusUnauthorized:                 "Authentication failed. Verify the credentials being used for the request.",
		http.StatusPaymentRequired:              "Payment required. Access to the requested resource requires payment.",
		http.StatusForbidden:                    "Invalid permissions. Verify the account has the proper permissions for the resource.",
		http.StatusNotFound:                     "Resource not found. Verify the URL path is correct.",
		http.StatusMethodNotAllowed:             "Method not allowed. The method specified is not allowed for the resource.",
		http.StatusNotAcceptable:                "Not acceptable. The server cannot produce a response matching the list of acceptable values.",
		http.StatusProxyAuthRequired:            "Proxy authentication required. You must authenticate with a proxy server before this request can be served.",
		http.StatusRequestTimeout:               "Request timeout. The server timed out waiting for the request.",
		http.StatusConflict:                     "Conflict. The request could not be processed because of conflict in the request.",
		http.StatusGone:                         "Gone. The resource requested is no longer available and will not be available again.",
		http.StatusLengthRequired:               "Length required. The request did not specify the length of its content, which is required by the requested resource.",
		http.StatusPreconditionFailed:           "Precondition failed. The server does not meet one of the preconditions specified in the request.",
		http.StatusRequestEntityTooLarge:        "Payload too large. The request is larger than the server is willing or able to process.",
		http.StatusRequestURITooLong:            "Request-URI too long. The URI provided was too long for the server to process.",
		http.StatusUnsupportedMediaType:         "Unsupported media type. The request entity has a media type which the server or resource does not support.",
		http.StatusRequestedRangeNotSatisfiable: "Requested range not satisfiable. The client has asked for a portion of the file, but the server cannot supply that portion.",
		http.StatusExpectationFailed:            "Expectation failed. The server cannot meet the requirements of the Expect request-header field.",
		http.StatusUnprocessableEntity:          "Unprocessable entity. The server understands the content type and syntax of the request but was unable to process the contained instructions.",
		http.StatusLocked:                       "Locked. The resource that is being accessed is locked.",
		http.StatusFailedDependency:             "Failed dependency. The request failed because it depended on another request and that request failed.",
		http.StatusUpgradeRequired:              "Upgrade required. The client should switch to a different protocol.",
		http.StatusPreconditionRequired:         "Precondition required. The server requires that the request be conditional.",
		http.StatusTooManyRequests:              "Too many requests. The user has sent too many requests in a given amount of time.",
		http.StatusRequestHeaderFieldsTooLarge:  "Request header fields too large. The server is unwilling to process the request because its header fields are too large.",
		http.StatusUnavailableForLegalReasons:   "Unavailable for legal reasons. The server is denying access to the resource as a consequence of a legal demand.",

		// Server error responses (500-599)
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
