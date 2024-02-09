// http_helpers.go
package httpclient

import (
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// ParseISO8601Date attempts to parse a string date in ISO 8601 format.
func ParseISO8601Date(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
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

// RedactSensitiveData redacts sensitive data if the HideSensitiveData flag is set to true.
func RedactSensitiveData(client *Client, key string, value string) string {
	if client.clientConfig.ClientOptions.HideSensitiveData {
		// Define sensitive data keys that should be redacted.
		sensitiveKeys := map[string]bool{
			"AccessToken": true,
			// Add more sensitive keys as necessary.
		}

		if _, found := sensitiveKeys[key]; found {
			return "REDACTED"
		}
	}
	return value
}
