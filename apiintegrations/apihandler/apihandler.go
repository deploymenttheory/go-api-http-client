// apiintegrations/apihandler/apihandler.go
package apihandler

import (
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/apiintegrations/jamfpro"
	"github.com/deploymenttheory/go-api-http-client/apiintegrations/msgraph"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// APIHandler is an interface for encoding, decoding, and implenting contexual api functions for different API implementations.
// It encapsulates behavior for encoding and decoding requests and responses.
type APIHandler interface {
	ConstructAPIResourceEndpoint(instanceName string, endpointPath string, log logger.Logger) string
	ConstructAPIAuthEndpoint(instanceName string, endpointPath string, log logger.Logger) string
	MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, error)
	HandleAPISuccessResponse(resp *http.Response, out interface{}, log logger.Logger) error
	HandleAPIErrorResponse(resp *http.Response, out interface{}, log logger.Logger) error
	GetContentTypeHeader(method string, log logger.Logger) string
	GetAcceptHeader() string
	GetDefaultBaseDomain() string
	GetOAuthTokenEndpoint() string
	GetBearerTokenEndpoint() string
	GetTokenRefreshEndpoint() string
	GetTokenInvalidateEndpoint() string
	GetAPIBearerTokenAuthenticationSupportStatus() bool
	GetAPIOAuthAuthenticationSupportStatus() bool
	GetAPIOAuthWithCertAuthenticationSupportStatus() bool
	GetAPIRequestHeaders(endpoint string) map[string]string // Provides standard headers required for making API requests.
}

// LoadAPIHandler returns an APIHandler based on the provided API type.
// 'apiType' parameter could be "jamf" or "graph" to specify which API handler to load.
func LoadAPIHandler(apiType string, log logger.Logger) (APIHandler, error) {
	var apiHandler APIHandler
	switch apiType {
	case "jamfpro":
		apiHandler = &jamfpro.JamfAPIHandler{
			Logger: log,
			// Initialize with necessary parameters
		}
		log.Info("API handler loaded successfully", zap.String("APIType", apiType))

	case "msgraph":
		apiHandler = &msgraph.GraphAPIHandler{
			// Initialize with necessary parameters
		}
		log.Info("API handler loaded successfully", zap.String("APIType", apiType))

	default:
		return nil, log.Error("Unsupported API type", zap.String("APIType", apiType))
	}

	return apiHandler, nil
}
