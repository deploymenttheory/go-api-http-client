// github_handler_constants.go
package github

// Endpoint constants represent the URL suffixes used for GitHub token interactions.
const (
	APIName                            = "github"                                       // APIName: represents the name of the API.
	DefaultBaseDomain                  = "api.github.com"                               // DefaultBaseDomain: represents the base domain for the github instance.
	OAuthTokenEndpoint                 = "github.com/login/oauth/access_token"          // OAuthTokenEndpoint: The endpoint to obtain an OAuth token.
	BearerTokenEndpoint                = ""                                             // BearerTokenEndpoint: The endpoint to obtain a bearer token.
	TokenRefreshEndpoint               = "github.com/login/oauth/access_token"          // TokenRefreshEndpoint: The endpoint to refresh an existing token.
	TokenInvalidateEndpoint            = "api.github.com/applications/:client_id/token" // TokenInvalidateEndpoint: The endpoint to invalidate an active token.
	BearerTokenAuthenticationSupport   = false                                          // BearerTokenAuthSuppport: A boolean to indicate if the API supports bearer token authentication.
	OAuthAuthenticationSupport         = true                                           // OAuthAuthSuppport: A boolean to indicate if the API supports OAuth authentication.
	OAuthWithCertAuthenticationSupport = true                                           // OAuthWithCertAuthSuppport: A boolean to indicate if the API supports OAuth with client certificate authentication.
)

// GetDefaultBaseDomain returns the default base domain used for constructing API URLs to the http client.
func (g *GitHubAPIHandler) GetDefaultBaseDomain() string {
	return DefaultBaseDomain
}

// GetOAuthTokenEndpoint returns the endpoint for obtaining an OAuth token. Used for constructing API URLs for the http client.
func (g *GitHubAPIHandler) GetOAuthTokenEndpoint() string {
	return OAuthTokenEndpoint
}

// GetBearerTokenEndpoint returns the endpoint for obtaining a bearer token. Used for constructing API URLs for the http client.
func (g *GitHubAPIHandler) GetBearerTokenEndpoint() string {
	return BearerTokenEndpoint
}

// GetTokenRefreshEndpoint returns the endpoint for refreshing an existing token. Used for constructing API URLs for the http client.
func (g *GitHubAPIHandler) GetTokenRefreshEndpoint() string {
	return TokenRefreshEndpoint
}

// GetTokenInvalidateEndpoint returns the endpoint for invalidating an active token. Used for constructing API URLs for the http client.
func (g *GitHubAPIHandler) GetTokenInvalidateEndpoint() string {
	return TokenInvalidateEndpoint
}

// GetAPIBearerTokenAuthenticationSupportStatus returns a boolean indicating if bearer token authentication is supported in the api handler.
func (g *GitHubAPIHandler) GetAPIBearerTokenAuthenticationSupportStatus() bool {
	return BearerTokenAuthenticationSupport
}

// GetAPIOAuthAuthenticationSupportStatus returns a boolean indicating if OAuth authentication is supported in the api handler.
func (g *GitHubAPIHandler) GetAPIOAuthAuthenticationSupportStatus() bool {
	return OAuthAuthenticationSupport
}

// GetAPIOAuthWithCertAuthenticationSupportStatus returns a boolean indicating if OAuth with client certificate authentication is supported in the api handler.
func (g *GitHubAPIHandler) GetAPIOAuthWithCertAuthenticationSupportStatus() bool {
	return OAuthWithCertAuthenticationSupport
}
