package httpclient

import (
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

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
