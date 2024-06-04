// apiintegrations/apihandler/apihandler.go
package httpclient

import (
	"net/http"
	"time"
)

// APIHandler is an interface for encoding, decoding, and implenting contexual api functions for different API implementations.
// It encapsulates behavior for encoding and decoding requests and responses.
type APIIntegration interface {
	Token() (string, error)
	Domain() string
	PrepRequestParamsForIntegration(req *http.Request, tokenRefreshBufferPeriod time.Duration) error

	// Utilities
	PrepRequestBodyForIntergration(body interface{}, method string, endpoint string) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string) ([]byte, string, error)

	// Info
	GetAuthMethodDescriptor() string
}
