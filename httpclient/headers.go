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
