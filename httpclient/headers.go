// headers/headers.go
package httpclient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/headers/redact"
	"github.com/deploymenttheory/go-api-http-client/version"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// SetAuthorization sets the Authorization header for the request.
func (c *Client) SetAuthorizationHeader(req *http.Request) {
	token := c.AuthToken
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}
	req.Header.Set("Authorization", token)
}

// SetContentType sets the Content-Type header for the request.
func SetContentType(req *http.Request, contentType string) {
	req.Header.Set("Content-Type", contentType)
}

// SetAccept sets the Accept header for the request.
func SetAccept(req *http.Request, acceptHeader string) {
	req.Header.Set("Accept", acceptHeader)
}

// SetUserAgent sets the User-Agent header for the request.
func SetUserAgent(req *http.Request, userAgent string) {
	req.Header.Set("User-Agent", userAgent)
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

// SetUserAgentHeader sets the User-Agent header for an HTTP request.
func SetUserAgentHeader() string {
	return fmt.Sprintf("%s/%s", version.UserAgentBase, version.SDKVersion)
}

// SetRequestHeaders sets the necessary HTTP headers for a given request using the APIHandler to determine the required headers.
func (c *Client) SetRequestHeaders(req *http.Request, endpoint string) {
	// Retrieve the standard headers required for the request
	standardHeaders := c.API.GetAPIRequestHeaders(endpoint)

	// Loop through the standard headers and set them on the request
	for header, value := range standardHeaders {
		if header == "Authorization" {
			// Set the Authorization header using the token
			c.SetAuthorizationHeader(req) // Ensure the token is correctly prefixed with "Bearer "
		} else if value != "" {
			req.Header.Set(header, value)
		}
	}
}

// LogHeaders prints all the current headers in the http.Request using the zap logger.
// It uses the RedactSensitiveHeaderData function to redact sensitive data based on the hideSensitiveData flag.
func (c *Client) LogHeaders(req *http.Request, hideSensitiveData bool) {
	if c.Logger.GetLogLevel() <= logger.LogLevelDebug {
		// Initialize a new Header to hold the potentially redacted headers
		redactedHeaders := http.Header{}

		for name, values := range req.Header {
			// Redact sensitive values
			if len(values) > 0 {
				// Use the first value for simplicity; adjust if multiple values per header are expected
				redactedValue := redact.RedactSensitiveHeaderData(hideSensitiveData, name, values[0])
				redactedHeaders.Set(name, redactedValue)
			}
		}

		// Convert the redacted headers to a string for logging
		headersStr := HeadersToString(redactedHeaders)

		// Log the redacted headers
		c.Logger.Debug("HTTP Request Headers", zap.String("Headers", headersStr))
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
	return strings.Join(headerStrings, "\n") // "\n" as seperator.
}

// CheckDeprecationHeader checks the response headers for the Deprecation header and logs a warning if present.
func CheckDeprecationHeader(resp *http.Response, log logger.Logger) {
	deprecationHeader := resp.Header.Get("Deprecation")
	if deprecationHeader != "" {

		log.Warn("API endpoint is deprecated",
			zap.String("Date", deprecationHeader),
			zap.String("Endpoint", resp.Request.URL.String()),
		)
	}
}
