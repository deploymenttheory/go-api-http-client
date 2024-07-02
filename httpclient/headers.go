// headers/headers.go
package httpclient

import (
	"net/http"
)

// CheckDeprecationHeader checks the response headers for the Deprecation header and logs a warning if present.
func (c *Client) CheckDeprecationHeader(resp *http.Response) {
	deprecationHeader := resp.Header.Get("Deprecation")
	if deprecationHeader != "" {
		c.Sugar.Warn("API endpoint is deprecated", deprecationHeader, resp.Request.URL.String())
	}
}
