// http_client_auth_token_management.go
package httpclient

import (
	"time"

	"go.uber.org/zap"
)

// TokenResponse represents the structure of a token response from the API.
type TokenResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

// ValidAuthTokenCheck checks if the current token is valid and not close to expiry.
// If the token is invalid, it tries to refresh it.
// It returns a boolean indicating the validity of the token and an error if there's a failure.
func (c *Client) ValidAuthTokenCheck() (bool, error) {
	log := c.Logger

	if c.Token == "" {
		log.Debug("No token found, attempting to obtain a new one")
		if c.AuthMethod == "bearer" {
			log.Info("Credential Match", zap.String("AuthMethod", c.AuthMethod))
			err := c.ObtainToken(log)
			if err != nil {
				return false, log.Error("Bearer token retrieval failed: invalid credentials. Verify accuracy.", zap.Error(err))
			}
		} else if c.AuthMethod == "oauth" {
			log.Info("Credential Match", zap.String("AuthMethod", c.AuthMethod))
			if err := c.ObtainOAuthToken(c.clientConfig.Auth); err != nil {
				return false, log.Error("OAuth token retrieval failed: invalid credentials. Verify accuracy.", zap.Error(err))
			}
		} else {
			return false, log.Error("No valid credentials provided. Unable to obtain a token", zap.String("authMethod", c.AuthMethod))
		}
	}

	if time.Until(c.Expiry) < c.clientConfig.ClientOptions.TokenRefreshBufferPeriod {
		var err error
		if c.clientConfig.Auth.Username != "" && c.clientConfig.Auth.Password != "" {
			err = c.RefreshToken(log)
		} else if c.clientConfig.Auth.ClientID != "" && c.clientConfig.Auth.ClientSecret != "" {
			err = c.ObtainOAuthToken(c.clientConfig.Auth)
		} else {
			return false, log.Error("Unknown auth method", zap.String("authMethod", c.AuthMethod))
		}

		if err != nil {
			return false, log.Error("Failed to refresh token", zap.Error(err))
		}
	}

	if time.Until(c.Expiry) < c.clientConfig.ClientOptions.TokenRefreshBufferPeriod {
		return false, log.Error(
			"Token lifetime setting less than buffer",
			zap.Duration("buffer_period", c.clientConfig.ClientOptions.TokenRefreshBufferPeriod),
			zap.Duration("time_until_expiry", time.Until(c.Expiry)),
		)
	}

	return true, nil
}
