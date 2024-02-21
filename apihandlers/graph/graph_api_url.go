// graph_api_url.go
package graph

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

// ConstructAPIAuthEndpoint constructs the full URL for a graph API auth endpoint path and logs the URL.
func (g *GraphAPIHandler) ConstructAPIAuthEndpoint(instanceName string, endpointPath string, log logger.Logger) string {
	urlBaseDomain := g.SetBaseDomain()
	url := fmt.Sprintf("https://%s%s%s", instanceName, urlBaseDomain, endpointPath)
	g.Logger.Debug(fmt.Sprintf("Constructed %s API authentication URL", APIName), zap.String("URL", url))
	return url
}
