// http_client_auth_token_management.go
package http_client

import (
	"fmt"
	"time"
)

// TokenResponse represents the structure of a token response from the API.
type TokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// ValidAuthTokenCheck checks if the current token is valid and not close to expiry.
// If the token is invalid or close to expiry, it tries to obtain a new token.
func (c *Client) ValidAuthTokenCheck() (bool, error) {
	if c.Token == "" || time.Until(c.Expiry) < c.config.TokenRefreshBufferPeriod {
		var oauthResp *OAuthResponse
		var err error

		switch c.AuthMethod {
		case "oauthApp":
			// Obtain token using OAuth App credentials
			oauthResp, err = c.ObtainOauthTokenWithApp(c.TenantID, c.OAuthCredentials.ClientID, c.OAuthCredentials.ClientSecret)

		case "oauthCertificate":
			// Obtain token using OAuth Certificate credentials
			oauthResp, err = c.ObtainOauthTokenWithCertificate(c.TenantID, c.OAuthCredentials.ClientID, c.OAuthCredentials.CertThumbprint, c.OAuthCredentials.CertificatePath)

		default:
			return false, fmt.Errorf("unknown auth method: %s", c.AuthMethod)
		}

		if err != nil {
			return false, fmt.Errorf("failed to obtain new token: %w", err)
		}

		// Update the token and expiry time if a new token was obtained
		if oauthResp != nil {
			c.Token = oauthResp.AccessToken
			expiresIn := time.Duration(oauthResp.ExpiresIn) * time.Second
			c.Expiry = time.Now().Add(expiresIn)
		}
	}

	return true, nil
}
