// apiintegrations/apihandler/apihandler.go
package httpclient

import (
	"net/http"
	"time"
)

// TODO comment
type APIIntegration interface {
	Domain() string
	PrepRequestParams(req *http.Request, tokenRefreshBufferPeriod time.Duration) error

	// Utilities
	PrepRequestBody(body interface{}, method string, endpoint string) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string) ([]byte, string, error)

	// Info
	GetAuthMethodDescriptor() string
}
