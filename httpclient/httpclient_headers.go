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
}

// NewHeaderManager creates a new instance of HeaderManager for a given http.Request, logger, and APIHandler.
func NewHeaderManager(req *http.Request, log logger.Logger, apiHandler APIHandler) *HeaderManager {
	return &HeaderManager{
		req:        req,
		log:        log,
		apiHandler: apiHandler, // Initialize with the provided APIHandler
	}
}

// Helper function to convert headers to string for logging
func HeadersToString(headers http.Header) string {
	var headerStrings []string
	for name, values := range headers {
		headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", name, strings.Join(values, ", ")))
	}
	return strings.Join(headerStrings, "; ")
}

// LogHeaders prints all the current headers in the http.Request using the zap logger.
func (h *HeaderManager) LogHeaders() {
	if h.log.GetLogLevel() <= logger.LogLevelDebug {
		headers := HeadersToString(h.req.Header)
		h.log.Debug("HTTP Request Headers", zap.String("Headers", headers))
	}
}

// SetAuthorization sets the Authorization header for the request.
func (h *HeaderManager) SetAuthorization(token string) {
	h.req.Header.Set("Authorization", "Bearer "+token)
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

// SetRequestHeaders sets the necessary HTTP headers for a given request. It configures the Authorization,
// Content-Type, and Accept headers based on the client's current token, the content type specified by the
// caller, and the preferred response formats defined by the APIHandler's GetAcceptHeader method.
// Additionally, it sets a User-Agent header to identify the client making the request.
// If debug logging is enabled, the function logs all set headers, with sensitive information such as the
// Authorization token being redacted for security purposes.
//
// Parameters:
// - req: The *http.Request object to which headers will be added. This request is prepared by the caller
// and passed to SetRequestHeaders for header configuration.
//
// - contentType: A string specifying the content type of the request, typically determined by the APIHandler's
// logic based on the request's nature and the endpoint being accessed.
//
// - acceptHeader: A string specifying the Accept header value, which is obtained from the APIHandler's
// GetAcceptHeader method. This header indicates the MIME types that the client can process, with preferences
// expressed through quality factors (q-values).
//
// - log: A logger.Logger instance used for logging header information when debug logging is enabled. The
// logger's level controls whether headers are logged, with logging occurring only at LogLevelDebug or lower.
//
// The function leverages the APIHandler interface to dynamically determine the appropriate Accept header,
// ensuring compatibility with the API's supported response formats. This approach allows for flexible and
// context-aware setting of request headers, facilitating effective communication with diverse APIs managed
// by different handlers, such as the JamfAPIHandler example provided in the api_handler.go file.
func (c *Client) SetRequestHeaders(req *http.Request, contentType, acceptHeader string, log logger.Logger) {
	// Set Headers
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", acceptHeader)
	req.Header.Set("User-Agent", GetUserAgentHeader())

	// Debug: Print request headers if debug logging is enabled
	if log.GetLogLevel() <= logger.LogLevelDebug {
		redactedAuthorization := RedactSensitiveData(c, "Authorization", req.Header.Get("Authorization"))
		log.Debug("HTTP Request Headers",
			zap.String("Authorization", redactedAuthorization),
			zap.String("Content-Type", req.Header.Get("Content-Type")),
			zap.String("Accept", req.Header.Get("Accept")),
			zap.String("User-Agent", req.Header.Get("User-Agent")),
		)
	}
}

// SetRequestHeaders sets the standard headers required for the API request.
func (h *HeaderManager) SetRequestHeaders(endpoint string) {
	// Use the APIHandler to get the standard request headers required for this endpoint
	standardHeaders := h.apiHandler.GetStandardRequestHeaders(endpoint)

	// Set each required header on the request
	for header, value := range standardHeaders {
		switch header {
		case "Authorization":
			// Assuming the token is set separately and not via GetStandardRequestHeaders
			h.SetAuthorization("Bearer " + value) // Adjust this according to how you manage tokens
		case "Content-Type", "Accept", "User-Agent":
			// Directly set the header if the value is provided, otherwise, it will be skipped
			if value != "" {
				h.req.Header.Set(header, value)
			}
		default:
			// Set custom headers that might be added by the APIHandler
			h.req.Header.Set(header, value)
		}
	}

	// Optionally log the headers for debugging
	h.LogHeaders()
}
