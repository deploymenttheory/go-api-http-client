// authenticationhandler/auth_token_management.go
package authenticationhandler

import (
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/apiintegrations/apihandler"
	"go.uber.org/zap"
)

// ValidAuthTokenCheck checks if the current token is valid and not close to expiry.
// If the token is invalid, it tries to refresh it.
// It returns a boolean indicating the validity of the token and an error if there's a failure.
func (h *AuthTokenHandler) ValidAuthTokenCheck(apiHandler apihandler.APIHandler, httpClient *http.Client, clientCredentials ClientCredentials, tokenRefreshBufferPeriod time.Duration) (bool, error) {
	if h.Token == "" {
		h.Logger.Debug("No token found, attempting to obtain a new one")
		var err error
		if h.AuthMethod == "bearer" {
			h.Logger.Info("Credential Match", zap.String("AuthMethod", h.AuthMethod))
			err = h.ObtainToken(apiHandler, httpClient, clientCredentials.Username, clientCredentials.Password)
		} else if h.AuthMethod == "oauth" {
			h.Logger.Info("Credential Match", zap.String("AuthMethod", h.AuthMethod))
			err = h.ObtainOAuthToken(apiHandler, httpClient, clientCredentials.ClientID, clientCredentials.ClientSecret)
		} else {
			return false, h.Logger.Error("No valid credentials provided. Unable to obtain a token", zap.String("AuthMethod", h.AuthMethod))
		}

		if err != nil {
			return false, err
		}
	}

	if time.Until(h.Expires) < tokenRefreshBufferPeriod {
		var err error
		if clientCredentials.Username != "" && clientCredentials.Password != "" {
			err = h.RefreshToken(apiHandler, httpClient)
		} else if clientCredentials.ClientID != "" && clientCredentials.ClientSecret != "" {
			err = h.ObtainOAuthToken(apiHandler, httpClient, clientCredentials.ClientID, clientCredentials.ClientSecret)
		} else {
			return false, h.Logger.Error("Unknown auth method", zap.String("authMethod", h.AuthMethod))
		}

		if err != nil {
			return false, h.Logger.Error("Failed to refresh token", zap.Error(err))
		}
	}

	if time.Until(h.Expires) < tokenRefreshBufferPeriod {
		return false, h.Logger.Error(
			"Token lifetime setting less than buffer",
			zap.Duration("buffer_period", tokenRefreshBufferPeriod),
			zap.Duration("time_until_expiry", time.Until(h.Expires)),
		)
	}

	return true, nil
}
