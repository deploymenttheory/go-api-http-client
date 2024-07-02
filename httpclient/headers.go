// headers/headers.go
package httpclient

import (
	"net/http"

	"go.uber.org/zap"
)

// CheckDeprecationHeader checks the response headers for the Deprecation header and logs a warning if present.
func (c *Client) CheckDeprecationHeader(resp *http.Response) {
	deprecationHeader := resp.Header.Get("Deprecation")
	if deprecationHeader != "" {
		c.Sugar.Warn("API endpoint is deprecated",
			zap.String("Date", deprecationHeader),
			zap.String("Endpoint", resp.Request.URL.String()),
		)
	}
}
