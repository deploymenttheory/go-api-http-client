// msgraph_api_url.go
package msgraph

import (
	"fmt"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// SetBaseDomain returns the appropriate base domain for URL construction.
// It uses j.OverrideBaseDomain if set, otherwise falls back to DefaultBaseDomain.
func (g *GraphAPIHandler) SetBaseDomain() string {
	if g.OverrideBaseDomain != "" {
		return g.OverrideBaseDomain
	}
	return DefaultBaseDomain
}

// ConstructAPIResourceEndpoint constructs the full URL for a graph API resource endpoint path and logs the URL.
func (g *GraphAPIHandler) ConstructAPIResourceEndpoint(instanceName string, endpointPath string, log logger.Logger) string {
	urlBaseDomain := g.SetBaseDomain()
	url := fmt.Sprintf("https://%s%s%s", instanceName, urlBaseDomain, endpointPath)
	g.Logger.Debug(fmt.Sprintf("Constructed %s API resource endpoint URL", APIName), zap.String("URL", url))
	return url
}

// ConstructAPIAuthEndpoint constructs the full URL for the Microsoft Graph API authentication endpoint.
// It uses the provided tenant name and endpoint path and logs the constructed URL.
func (g *GraphAPIHandler) ConstructAPIAuthEndpoint(tenantName string, endpointPath string, log logger.Logger) string {
	// The base URL for the Microsoft Graph API authentication endpoint.
	const baseURL = "https://login.microsoftonline.com"

	// Construct the full URL by combining the base URL, tenant name, and endpoint path.
	url := fmt.Sprintf("%s/%s%s", baseURL, tenantName, endpointPath)

	// Log the constructed URL for debugging purposes.
	log.Debug(fmt.Sprintf("Constructed Microsoft Graph API authentication URL"), zap.String("URL", url))

	return url
}
