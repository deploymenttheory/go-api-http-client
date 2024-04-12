// jamfpro_api_url.go
package jamfpro

import (
	"fmt"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// SetBaseDomain returns the appropriate base domain for URL construction.
// It uses j.OverrideBaseDomain if set, otherwise falls back to DefaultBaseDomain.
func (j *JamfAPIHandler) SetBaseDomain() string {
	if j.OverrideBaseDomain != "" {
		return j.OverrideBaseDomain
	}
	return DefaultBaseDomain
}

// ConstructAPIResourceEndpoint constructs the full URL for a Jamf API resource endpoint path and logs the URL.
// It uses the instance name to construct the full URL.
func (j *JamfAPIHandler) ConstructAPIResourceEndpoint(endpointPath string, log logger.Logger) string {
	urlBaseDomain := j.SetBaseDomain()
	url := fmt.Sprintf("https://%s%s%s", j.InstanceName, urlBaseDomain, endpointPath)
	j.Logger.Debug(fmt.Sprintf("Constructed %s API resource endpoint URL", APIName), zap.String("URL", url))
	return url
}

// ConstructAPIAuthEndpoint constructs the full URL for a Jamf API auth endpoint path and logs the URL.
// It uses the instance name to construct the full URL.
func (j *JamfAPIHandler) ConstructAPIAuthEndpoint(endpointPath string, log logger.Logger) string {
	urlBaseDomain := j.SetBaseDomain()
	url := fmt.Sprintf("https://%s%s%s", j.InstanceName, urlBaseDomain, endpointPath)
	j.Logger.Debug(fmt.Sprintf("Constructed %s API authentication URL", APIName), zap.String("URL", url))
	return url
}
