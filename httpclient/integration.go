// apiintegrations/apihandler/apihandler.go
package httpclient

import (
	"net/http"
)

// TODO comment
type APIIntegration interface {
	Domain() string
	GetAuthMethodDescriptor() string
	CheckRefreshToken() error
	PrepRequestParamsAndAuth(req *http.Request) error
	PrepRequestBody(body interface{}, method string, endpoint string) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string) ([]byte, string, error)
	GetSessionCookies() ([]*http.Cookie, error)
}
