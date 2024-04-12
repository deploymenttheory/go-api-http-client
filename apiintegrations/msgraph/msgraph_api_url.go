// msgraph_api_url.go
package msgraph

import (
	"fmt"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// SetBaseDomain returns the appropriate base domain for URL construction. It uses DefaultBaseDomain constant.
func (g *GraphAPIHandler) SetBaseDomain() string {
	return DefaultBaseDomain
}

// ConstructAPIResourceEndpoint constructs the full URL for a graph API resource endpoint path and logs the URL.
// It uses the base domain to construct the full URL.
func (g *GraphAPIHandler) ConstructAPIResourceEndpoint(endpointPath string, log logger.Logger) string {
	urlBaseDomain := g.SetBaseDomain()
	url := fmt.Sprintf("https://%s%s", urlBaseDomain, endpointPath)
	g.Logger.Debug(fmt.Sprintf("Constructed %s API resource endpoint URL", APIName), zap.String("URL", url))
	return url
}

// ConstructAPIAuthEndpoint constructs the full URL for the Microsoft Graph API authentication endpoint.
// It uses the tenant ID to construct the full URL.
func (g *GraphAPIHandler) ConstructAPIAuthEndpoint(endpointPath string, log logger.Logger) string {
	// The base URL for the Microsoft Graph API authentication endpoint.
	const baseURL = "https://login.microsoftonline.com"

	// Construct the full URL by combining the base URL, tenant ID, and endpoint path.
	url := fmt.Sprintf("%s/%s%s", baseURL, g.TenantID, endpointPath)

	// Log the constructed URL for debugging purposes.
	log.Debug("constructed Microsoft Graph API authentication URL", zap.String("URL", url))

	return url
}
