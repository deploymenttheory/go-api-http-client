package jamfpro

import "github.com/deploymenttheory/go-api-http-client/logger"

// JamfAPIHandler implements the APIHandler interface for the Jamf Pro API.
type JamfAPIHandler struct {
	OverrideBaseDomain string        // OverrideBaseDomain is used to override the base domain for URL construction.
	InstanceName       string        // InstanceName is the name of the Jamf instance.
	Logger             logger.Logger // Logger is the structured logger used for logging.
}
