// headers/headers.go
package httpclient

import (
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

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

// TODO review the need for headers below. Do they need to be in the Integration?

// SetCacheControlHeader sets the Cache-Control header for an HTTP request.
// This header specifies directives for caching mechanisms in requests and responses.
// func SetCacheControlHeader(req *http.Request, cacheControlValue string) {
// 	req.Header.Set("Cache-Control", cacheControlValue)
// }

// SetConditionalHeaders sets the If-Modified-Since and If-None-Match headers for an HTTP request.
// These headers make a request conditional to ask the server to return content only if it has changed.
// func SetConditionalHeaders(req *http.Request, ifModifiedSince, ifNoneMatch string) {
// 	if ifModifiedSince != "" {
// 		req.Header.Set("If-Modified-Since", ifModifiedSince)
// 	}
// 	if ifNoneMatch != "" {
// 		req.Header.Set("If-None-Match", ifNoneMatch)
// 	}
// }

// SetAcceptEncodingHeader sets the Accept-Encoding header for an HTTP request.
// This header indicates the type of encoding (e.g., gzip) the client can handle.
// func SetAcceptEncodingHeader(req *http.Request, acceptEncodingValue string) {
// 	req.Header.Set("Accept-Encoding", acceptEncodingValue)
// }

// SetRefererHeader sets the Referer header for an HTTP request.
// This header indicates the address of the previous web page from which a link was followed.
// func SetRefererHeader(req *http.Request, refererValue string) {
// 	req.Header.Set("Referer", refererValue)
// }

// SetXForwardedForHeader sets the X-Forwarded-For header for an HTTP request.
// This header is used to identify the originating IP address of a client connecting through a proxy.
// func SetXForwardedForHeader(req *http.Request, xForwardedForValue string) {
// 	req.Header.Set("X-Forwarded-For", xForwardedForValue)
// }

// LogHeaders prints all the current headers in the http.Request using the zap logger.
// It uses the RedactSensitiveHeaderData function to redact sensitive data based on the hideSensitiveData flag.
// func (c *Client) LogHeaders(req *http.Request, hideSensitiveData bool) {
// 	if c.Logger.GetLogLevel() <= logger.LogLevelDebug {
// 		redactedHeaders := http.Header{}

// 		for name, values := range req.Header {
// 			if len(values) > 0 {
// 				redactedValue := redact.RedactSensitiveHeaderData(hideSensitiveData, name, values[0])
// 				redactedHeaders.Set(name, redactedValue)
// 			}
// 		}

// 		headersStr := HeadersToString(redactedHeaders)

// 		c.Logger.Debug("HTTP Request Headers", zap.String("Headers", headersStr))
// 	}
// }

// HeadersToString converts a http.Header to a string for logging,
// with each header on a new line for readability.
// func HeadersToString(headers http.Header) string {
// 	var headerStrings []string
// 	for name, values := range headers {
// 		valueStr := strings.Join(values, ", ")
// 		headerStrings = append(headerStrings, fmt.Sprintf("%s: %s", name, valueStr))
// 	}
// 	return strings.Join(headerStrings, "\n") // "\n" as seperator.
// }
