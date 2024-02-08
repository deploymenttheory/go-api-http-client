package httpclient

import (
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/internal/logger"
)

// LogHTTPHeaders logs the HTTP headers of an HTTP request or response, with an option to hide sensitive information like the token in secure mode.
func LogHTTPHeaders(log logger.Logger, headers http.Header, secureMode bool) {
	var keysAndValues []interface{}
	if secureMode {
		for key, values := range headers {
			if key != "Authorization" { // Exclude the token header
				// Assuming each header has a single value for simplicity
				keysAndValues = append(keysAndValues, key, values[0])
			}
		}
	} else {
		for key, values := range headers {
			// Assuming each header has a single value for simplicity
			keysAndValues = append(keysAndValues, key, values[0])
		}
	}

	// Log the headers using the logger from the httpclient package
	if len(keysAndValues) > 0 {
		log.Debug("HTTP Headers", keysAndValues...)
	}
}
