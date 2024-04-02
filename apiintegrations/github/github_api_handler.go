// github_api_handler.go
package github

import "github.com/deploymenttheory/go-api-http-client/logger"

// GitHubHandler implements the APIHandler interface for the GitHub API.
type GitHubAPIHandler struct {
	OverrideBaseDomain string        // OverrideBaseDomain is used to override the base domain for URL construction.
	InstanceName       string        // InstanceName is the name of the GitHub instance.
	Logger             logger.Logger // Logger is the structured logger used for logging.
}
