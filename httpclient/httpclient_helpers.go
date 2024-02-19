// http_helpers.go
package httpclient

import (
	"time"
)

// ParseISO8601Date attempts to parse a string date in ISO 8601 format.
func ParseISO8601Date(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}

// RedactSensitiveData redacts sensitive data if the HideSensitiveData flag is set to true.
func RedactSensitiveData(client *Client, key string, value string) string {
	if client.clientConfig.ClientOptions.HideSensitiveData {
		// Define sensitive data keys that should be redacted.
		sensitiveKeys := map[string]bool{
			"AccessToken":   true,
			"Authorization": true,
		}

		if _, found := sensitiveKeys[key]; found {
			return "REDACTED"
		}
	}
	return value
}
