// http_client_auth_token_management.go
package httpclient

import (
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
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
func (c *Client) ValidAuthTokenCheck(log logger.Logger) (bool, error) {
	if c.Token == "" {
		c.Logger.Debug("No token found, attempting to obtain a new one")
		if c.AuthMethod == "bearer" {
			c.Logger.Info("Credential Match", zap.String("AuthMethod", c.AuthMethod))
			err := c.ObtainToken(log)
			if err != nil {
				return false, c.Logger.Error("Failed to obtain bearer token", zap.Error(err))
			}
		} else if c.AuthMethod == "oauth" {
			c.Logger.Info("Credential Match", zap.String("AuthMethod", c.AuthMethod))
			if err := c.ObtainOAuthToken(c.config.Auth, log); err != nil {
				return false, c.Logger.Error("Failed to obtain OAuth token", zap.Error(err))
			}
		} else {
			return false, c.Logger.Error("No valid credentials provided. Unable to obtain a token", zap.String("authMethod", c.AuthMethod))
		}
	}

	if time.Until(c.Expiry) < c.config.TokenRefreshBufferPeriod {
		var err error
		if c.BearerTokenAuthCredentials.Username != "" && c.BearerTokenAuthCredentials.Password != "" {
			err = c.RefreshToken(log)
		} else if c.OAuthCredentials.ClientID != "" && c.OAuthCredentials.ClientSecret != "" {
			err = c.ObtainOAuthToken(c.config.Auth, log)
		} else {
			return false, c.Logger.Error("Unknown auth method", zap.String("authMethod", c.AuthMethod))
		}

		if err != nil {
			return false, c.Logger.Error("Failed to refresh token", zap.Error(err))
		}
	}

	if time.Until(c.Expiry) < c.config.TokenRefreshBufferPeriod {
		return false, c.Logger.Error(
			"Token lifetime setting less than buffer",
			zap.Duration("buffer_period", c.config.TokenRefreshBufferPeriod),
			zap.Duration("time_until_expiry", time.Until(c.Expiry)),
		)
	}

	return true, nil
}
