// apiintegrations/apihandler/apihandler.go
package httpclient

import (
	"net/http"
)

// APIHandler is an interface for encoding, decoding, and implenting contexual api functions for different API implementations.
// It encapsulates behavior for encoding and decoding requests and responses.
type APIIntegration interface {
	Token() (string, error)
	Domain() string
	SetRequestHeaders(req *http.Request)

	// Utilities
	MarshalRequest(body interface{}, method string, endpoint string) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string) ([]byte, string, error)
	GetContentTypeHeader(method string) string

	// Info
	AuthMethodDescriptor() string
}
