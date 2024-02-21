// graph_api_handler_constants.go
package graph

// Endpoint constants represent the URL suffixes used for graph API token interactions.
const (
	APIName                            = "microsoft graph"     // APIName: represents the name of the API.
	DefaultBaseDomain                  = "graph.microsoft.com" // DefaultBaseDomain: represents the base domain for the graph instance.
	OAuthTokenEndpoint                 = "/oauth2/v2.0/token"  // OAuthTokenEndpoint: The endpoint to obtain an OAuth token.
	BearerTokenEndpoint                = ""                    // BearerTokenEndpoint: The endpoint to obtain a bearer token.
	TokenRefreshEndpoint               = "graph.microsoft.com" // TokenRefreshEndpoint: The endpoint to refresh an existing token.
	TokenInvalidateEndpoint            = "graph.microsoft.com" // TokenInvalidateEndpoint: The endpoint to invalidate an active token.
	BearerTokenAuthenticationSupport   = true                  // BearerTokenAuthSuppport: A boolean to indicate if the API supports bearer token authentication.
	OAuthAuthenticationSupport         = true                  // OAuthAuthSuppport: A boolean to indicate if the API supports OAuth authentication.
	OAuthWithCertAuthenticationSupport = true                  // OAuthWithCertAuthSuppport: A boolean to indicate if the API supports OAuth with client certificate authentication.
)

// GetDefaultBaseDomain returns the default base domain used for constructing API URLs to the http client.
func (g *GraphAPIHandler) GetDefaultBaseDomain() string {
	return DefaultBaseDomain
}

// GetOAuthTokenEndpoint returns the endpoint for obtaining an OAuth token. Used for constructing API URLs for the http client.
func (g *GraphAPIHandler) GetOAuthTokenEndpoint() string {
	return OAuthTokenEndpoint
}

// GetBearerTokenEndpoint returns the endpoint for obtaining a bearer token. Used for constructing API URLs for the http client.
func (g *GraphAPIHandler) GetBearerTokenEndpoint() string {
	return BearerTokenEndpoint
}

// GetTokenRefreshEndpoint returns the endpoint for refreshing an existing token. Used for constructing API URLs for the http client.
func (g *GraphAPIHandler) GetTokenRefreshEndpoint() string {
	return TokenRefreshEndpoint
}

// GetTokenInvalidateEndpoint returns the endpoint for invalidating an active token. Used for constructing API URLs for the http client.
func (g *GraphAPIHandler) GetTokenInvalidateEndpoint() string {
	return TokenInvalidateEndpoint
}

// GetAPIBearerTokenAuthenticationSupportStatus returns a boolean indicating if bearer token authentication is supported in the api handler.
func (g *GraphAPIHandler) GetAPIBearerTokenAuthenticationSupportStatus() bool {
	return BearerTokenAuthenticationSupport
}

// GetAPIOAuthAuthenticationSupportStatus returns a boolean indicating if OAuth authentication is supported in the api handler.
func (g *GraphAPIHandler) GetAPIOAuthAuthenticationSupportStatus() bool {
	return OAuthAuthenticationSupport
}

// GetAPIOAuthWithCertAuthenticationSupportStatus returns a boolean indicating if OAuth with client certificate authentication is supported in the api handler.
func (g *GraphAPIHandler) GetAPIOAuthWithCertAuthenticationSupportStatus() bool {
	return OAuthWithCertAuthenticationSupport
}
