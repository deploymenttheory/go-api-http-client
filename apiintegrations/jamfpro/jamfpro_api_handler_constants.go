package jamfpro

// Endpoint constants represent the URL suffixes used for Jamf API token interactions.
const (
	APIName                            = "jamf pro"                      // APIName: represents the name of the API.
	DefaultBaseDomain                  = ".jamfcloud.com"                // DefaultBaseDomain: represents the base domain for the jamf instance.
	OAuthTokenEndpoint                 = "/api/oauth/token"              // OAuthTokenEndpoint: The endpoint to obtain an OAuth token.
	OAuthTokenScope                    = ""                              // OAuthTokenScope: Not used for Jamf.
	BearerTokenEndpoint                = "/api/v1/auth/token"            // BearerTokenEndpoint: The endpoint to obtain a bearer token.
	TokenRefreshEndpoint               = "/api/v1/auth/keep-alive"       // TokenRefreshEndpoint: The endpoint to refresh an existing token.
	TokenInvalidateEndpoint            = "/api/v1/auth/invalidate-token" // TokenInvalidateEndpoint: The endpoint to invalidate an active token.
	BearerTokenAuthenticationSupport   = true                            // BearerTokenAuthSuppport: A boolean to indicate if the API supports bearer token authentication.
	OAuthAuthenticationSupport         = true                            // OAuthAuthSuppport: A boolean to indicate if the API supports OAuth authentication.
	OAuthWithCertAuthenticationSupport = false                           // OAuthWithCertAuthSuppport: A boolean to indicate if the API supports OAuth with client certificate authentication.
)

// GetDefaultBaseDomain returns the default base domain used for constructing API URLs to the http client.
func (j *JamfAPIHandler) GetDefaultBaseDomain() string {
	return DefaultBaseDomain
}

// GetOAuthTokenEndpoint returns the endpoint for obtaining an OAuth token. Used for constructing API URLs for the http client.
func (j *JamfAPIHandler) GetOAuthTokenEndpoint() string {
	return OAuthTokenEndpoint
}

// GetOAuthTokenScope returns the scope for the OAuth token scope
func (j *JamfAPIHandler) GetOAuthTokenScope() string {
	return OAuthTokenScope
}

// GetBearerTokenEndpoint returns the endpoint for obtaining a bearer token. Used for constructing API URLs for the http client.
func (j *JamfAPIHandler) GetBearerTokenEndpoint() string {
	return BearerTokenEndpoint
}

// GetTokenRefreshEndpoint returns the endpoint for refreshing an existing token. Used for constructing API URLs for the http client.
func (j *JamfAPIHandler) GetTokenRefreshEndpoint() string {
	return TokenRefreshEndpoint
}

// GetTokenInvalidateEndpoint returns the endpoint for invalidating an active token. Used for constructing API URLs for the http client.
func (j *JamfAPIHandler) GetTokenInvalidateEndpoint() string {
	return TokenInvalidateEndpoint
}

// GetAPIBearerTokenAuthenticationSupportStatus returns a boolean indicating if bearer token authentication is supported in the api handler.
func (j *JamfAPIHandler) GetAPIBearerTokenAuthenticationSupportStatus() bool {
	return BearerTokenAuthenticationSupport
}

// GetAPIOAuthAuthenticationSupportStatus returns a boolean indicating if OAuth authentication is supported in the api handler.
func (j *JamfAPIHandler) GetAPIOAuthAuthenticationSupportStatus() bool {
	return OAuthAuthenticationSupport
}

// GetAPIOAuthWithCertAuthenticationSupportStatus returns a boolean indicating if OAuth with client certificate authentication is supported in the api handler.
func (j *JamfAPIHandler) GetAPIOAuthWithCertAuthenticationSupportStatus() bool {
	return OAuthWithCertAuthenticationSupport
}
