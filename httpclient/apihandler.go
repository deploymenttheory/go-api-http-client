// apiintegrations/apihandler/apihandler.go
package httpclient

import (
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
)

// APIHandler is an interface for encoding, decoding, and implenting contexual api functions for different API implementations.
// It encapsulates behavior for encoding and decoding requests and responses.
type APIHandler interface {
	// Auth Endpoints
	GetBearerAuthEndpoint(log logger.Logger) string
	GetOAuthEndpoint(log logger.Logger) string

	// Resource Endpoints
	ConstructAPIResourceEndpoint(endpointPath string, log logger.Logger) string

	// Headers
	GetAcceptHeader() string
	GetAPIRequestHeaders(endpoint string) map[string]string // Provides standard headers required for making API requests.

	// Utilities
	MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, error)
	GetContentTypeHeader(method string, log logger.Logger) string

	// Not sure if we need all of these yet.
	GetDefaultBaseDomain() string
	GetOAuthTokenEndpoint() string
	GetOAuthTokenScope() string
	GetBearerTokenEndpoint() string
	GetTokenRefreshEndpoint() string
	GetTokenInvalidateEndpoint() string
	GetAPIBearerTokenAuthenticationSupportStatus() bool
	GetAPIOAuthAuthenticationSupportStatus() bool
	GetAPIOAuthWithCertAuthenticationSupportStatus() bool
	GetTokenResponseStruct()
}

type TokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}
