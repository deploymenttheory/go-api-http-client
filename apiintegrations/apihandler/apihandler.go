// apiintegrations/apihandler/apihandler.go
package apihandler

import (
	"github.com/deploymenttheory/go-api-http-client/apiintegrations/jamfpro"
	"github.com/deploymenttheory/go-api-http-client/apiintegrations/msgraph"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// APIHandler is an interface for encoding, decoding, and implenting contexual api functions for different API implementations.
// It encapsulates behavior for encoding and decoding requests and responses.
type APIHandler interface {
	ConstructAPIResourceEndpoint(APIResourceEndpointIdentifier string, endpointPath string, log logger.Logger) string
	ConstructAPIAuthEndpoint(APIAuthEndpointIdentifier string, endpointPath string, log logger.Logger) string
	MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, error)
	GetContentTypeHeader(method string, log logger.Logger) string
	GetAcceptHeader() string
	GetDefaultBaseDomain() string
	GetOAuthTokenEndpoint() string
	GetOAuthTokenScope() string
	GetBearerTokenEndpoint() string
	GetTokenRefreshEndpoint() string
	GetTokenInvalidateEndpoint() string
	GetAPIBearerTokenAuthenticationSupportStatus() bool
	GetAPIOAuthAuthenticationSupportStatus() bool
	GetAPIOAuthWithCertAuthenticationSupportStatus() bool
	GetAPIRequestHeaders(endpoint string) map[string]string // Provides standard headers required for making API requests.
}

// Modify the function signature to accept instanceName, tenantID, and tenantName.
func LoadAPIHandler(apiType, instanceName, tenantID, tenantName string, log logger.Logger) (APIHandler, error) {
	var apiHandler APIHandler
	switch apiType {
	case "jamfpro":
		apiHandler = &jamfpro.JamfAPIHandler{
			Logger:       log,
			InstanceName: instanceName, // Assuming you add InstanceName field to JamfAPIHandler
		}
		log.Info("Jamf Pro API handler loaded successfully", zap.String("APIType", apiType), zap.String("InstanceName", instanceName))

	case "msgraph":
		apiHandler = &msgraph.GraphAPIHandler{
			Logger:     log,
			TenantID:   tenantID,   // Assuming you add TenantID field to GraphAPIHandler
			TenantName: tenantName, // Assuming you add TenantName field to GraphAPIHandler
		}
		log.Info("Microsoft Graph API handler loaded successfully", zap.String("APIType", apiType), zap.String("TenantID", tenantID), zap.String("TenantName", tenantName))

	default:
		return nil, log.Error("Unsupported API type", zap.String("APIType", apiType))
	}

	return apiHandler, nil
}
