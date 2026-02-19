// httpclient/integration.go
package httpclient

import (
	"net/http"
)

// APIIntegration is an interface that defines the methods required for an API integration. These are obtained from go-api-http-client-integrations.
// The methods defined in this interface are used by the HTTP client to authenticate and prepare requests for the API.
type APIIntegration interface {
	GetFQDN() string
	ConstructURL(endpoint string) string
	GetAuthMethodDescriptor() string
	CheckRefreshToken() error
	PrepRequestParamsAndAuth(req *http.Request) error
	PrepRequestBody(body any, method string, endpoint string) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string) ([]byte, string, error)
	GetSessionCookies() ([]*http.Cookie, error)
}
