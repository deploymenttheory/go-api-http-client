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
