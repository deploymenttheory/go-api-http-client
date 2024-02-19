package httpclient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// HeaderManager is responsible for managing and setting headers on HTTP requests.
type HeaderManager struct {
	req        *http.Request
	log        logger.Logger
	apiHandler APIHandler
	token      string
}

// NewHeaderManager creates a new instance of HeaderManager for a given http.Request, logger, and APIHandler.
func NewHeaderManager(req *http.Request, log logger.Logger, apiHandler APIHandler, token string) *HeaderManager {
	return &HeaderManager{
		req:        req,
		log:        log,
		apiHandler: apiHandler, // Initialize with the provided APIHandler
		token:      token,      // Initialize with the provided token
	}
}

// HeadersToString converts a http.Header to a string for logging,
// with each header on a new line for readability.
func HeadersToString(headers http.Header) string {
	var headerStrings []string
	for name, values := range headers {
		// Join all values for the header with a comma, as per HTTP standard
		valueStr := strings.Join(values, ", ")
		headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", name, valueStr))
	}
	return strings.Join(headerStrings, "\n") // Use "\n" for new line separation in logs
}

// LogHeaders prints all the current headers in the http.Request using the zap logger.
// It uses the RedactSensitiveData function to redact sensitive data if required.
func (h *HeaderManager) LogHeaders(client *Client) {
	if h.log.GetLogLevel() <= logger.LogLevelDebug {
		// Initialize a new Header to hold the potentially redacted headers
		redactedHeaders := http.Header{}

		for name, values := range h.req.Header {
			// Redact sensitive values
			if len(values) > 0 {
				// Use the first value for simplicity; adjust if multiple values per header are expected
				redactedValue := RedactSensitiveData(client, name, values[0])
				redactedHeaders.Set(name, redactedValue)
			}
		}

		// Convert the redacted headers to a string for logging
		headersStr := HeadersToString(redactedHeaders)

		// Log the redacted headers
		h.log.Debug("HTTP Request Headers", zap.String("Headers", headersStr))
	}
}

// SetAuthorization sets the Authorization header for the request.
func (h *HeaderManager) SetAuthorization(token string) {
	// Ensure the token is prefixed with "Bearer " only once
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}
	h.req.Header.Set("Authorization", token)
}

// SetContentType sets the Content-Type header for the request.
func (h *HeaderManager) SetContentType(contentType string) {
	h.req.Header.Set("Content-Type", contentType)
}

// SetAccept sets the Accept header for the request.
func (h *HeaderManager) SetAccept(acceptHeader string) {
	h.req.Header.Set("Accept", acceptHeader)
}

// SetUserAgent sets the User-Agent header for the request.
func (h *HeaderManager) SetUserAgent(userAgent string) {
	h.req.Header.Set("User-Agent", userAgent)
}

// SetCacheControlHeader sets the Cache-Control header for an HTTP request.
// This header specifies directives for caching mechanisms in requests and responses.
func SetCacheControlHeader(req *http.Request, cacheControlValue string) {
	req.Header.Set("Cache-Control", cacheControlValue)
}

// SetConditionalHeaders sets the If-Modified-Since and If-None-Match headers for an HTTP request.
// These headers make a request conditional to ask the server to return content only if it has changed.
func SetConditionalHeaders(req *http.Request, ifModifiedSince, ifNoneMatch string) {
	if ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}
	if ifNoneMatch != "" {
		req.Header.Set("If-None-Match", ifNoneMatch)
	}
}

// SetAcceptEncodingHeader sets the Accept-Encoding header for an HTTP request.
// This header indicates the type of encoding (e.g., gzip) the client can handle.
func SetAcceptEncodingHeader(req *http.Request, acceptEncodingValue string) {
	req.Header.Set("Accept-Encoding", acceptEncodingValue)
}

// SetRefererHeader sets the Referer header for an HTTP request.
// This header indicates the address of the previous web page from which a link was followed.
func SetRefererHeader(req *http.Request, refererValue string) {
	req.Header.Set("Referer", refererValue)
}

// SetXForwardedForHeader sets the X-Forwarded-For header for an HTTP request.
// This header is used to identify the originating IP address of a client connecting through a proxy.
func SetXForwardedForHeader(req *http.Request, xForwardedForValue string) {
	req.Header.Set("X-Forwarded-For", xForwardedForValue)
}

// SetCustomHeader sets a custom header for an HTTP request.
// This function allows setting arbitrary headers for specialized API requirements.
func SetCustomHeader(req *http.Request, headerName, headerValue string) {
	req.Header.Set(headerName, headerValue)
}

// SetRequestHeaders sets the necessary HTTP headers for a given request using the APIHandler to determine the required headers.
func (h *HeaderManager) SetRequestHeaders(endpoint string) {
	// Retrieve the standard headers required for the request
	standardHeaders := h.apiHandler.GetAPIRequestHeaders(endpoint)

	// Loop through the standard headers and set them on the request
	for header, value := range standardHeaders {
		if header == "Authorization" {
			// Set the Authorization header using the token
			h.SetAuthorization(h.token) // Ensure the token is correctly prefixed with "Bearer "
		} else if value != "" {
			h.req.Header.Set(header, value)
		}
	}
}
