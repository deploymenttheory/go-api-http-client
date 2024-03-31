// msgraph_api_handler.go
package msgraph

import "github.com/deploymenttheory/go-api-http-client/logger"

// GraphAPIHandler implements the APIHandler interface for the graph Pro API.
type GraphAPIHandler struct {
	OverrideBaseDomain string        // OverrideBaseDomain is used to override the base domain for URL construction.
	InstanceName       string        // InstanceName is the name of the graph instance.
	Logger             logger.Logger // Logger is the structured logger used for logging.
}
