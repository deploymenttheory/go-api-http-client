package jamfpro

import "github.com/deploymenttheory/go-api-http-client/logger"

// EndpointConfig is a struct that holds configuration details for a specific API endpoint.
// It includes what type of content it can accept and what content type it should send.
type EndpointConfig struct {
	Accept      string  `json:"accept"`       // Accept specifies the MIME type the endpoint can handle in responses.
	ContentType *string `json:"content_type"` // ContentType, if not nil, specifies the MIME type to set for requests sent to the endpoint. A pointer is used to distinguish between a missing field and an empty string.
}

// JamfAPIHandler implements the APIHandler interface for the Jamf Pro API.
type JamfAPIHandler struct {
	OverrideBaseDomain string        // OverrideBaseDomain is used to override the base domain for URL construction.
	InstanceName       string        // InstanceName is the name of the Jamf instance.
	Logger             logger.Logger // Logger is the structured logger used for logging.
}
