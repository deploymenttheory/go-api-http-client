// authenticationhandler/tokenmanager.go
package authenticationhandler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/apiintegrations/apihandler"
	"go.uber.org/zap"
)

// CheckAndRefreshAuthToken checks the token's validity and refreshes it if necessary.
// It returns true if the token is valid post any required operations and false with an error otherwise.
func (h *AuthTokenHandler) CheckAndRefreshAuthToken(apiHandler apihandler.APIHandler, httpClient *http.Client, clientCredentials ClientCredentials, tokenRefreshBufferPeriod time.Duration) (bool, error) {
	const maxConsecutiveRefreshAttempts = 10
	refreshAttempts := 0

	for !h.isTokenValid(tokenRefreshBufferPeriod) {
		h.Logger.Debug("Token found to be invalid or close to expiry, handling token acquisition or refresh.")
		if err := h.obtainNewToken(apiHandler, httpClient, clientCredentials); err != nil {
			h.Logger.Error("Failed to obtain new token", zap.Error(err))
			return false, err
		}

		refreshAttempts++
		if refreshAttempts >= maxConsecutiveRefreshAttempts {
			return false, fmt.Errorf("exceeded maximum consecutive token refresh attempts (%d): token lifetime is likely too short", maxConsecutiveRefreshAttempts)
		}
	}

	if err := h.refreshTokenIfNeeded(apiHandler, httpClient, clientCredentials, tokenRefreshBufferPeriod); err != nil {
		h.Logger.Error("Failed to refresh token", zap.Error(err))
		return false, err
	}

	isValid := h.isTokenValid(tokenRefreshBufferPeriod)
	h.Logger.Info("Authentication token status check completed", zap.Bool("IsTokenValid", isValid))
	return isValid, nil
}

// isTokenValid checks if the current token is non-empty and not about to expire.
// It considers a token valid if it exists and the time until its expiration is greater than the provided buffer period.
func (h *AuthTokenHandler) isTokenValid(tokenRefreshBufferPeriod time.Duration) bool {
	isValid := h.Token != "" && time.Until(h.Expires) >= tokenRefreshBufferPeriod
	h.Logger.Debug("Checking token validity", zap.Bool("IsValid", isValid), zap.Duration("TimeUntilExpiry", time.Until(h.Expires)))
	return isValid
}

// obtainNewToken acquires a new token using the credentials provided.
// It handles different authentication methods based on the AuthMethod setting.
func (h *AuthTokenHandler) obtainNewToken(apiHandler apihandler.APIHandler, httpClient *http.Client, clientCredentials ClientCredentials) error {
	var err error
	if h.AuthMethod == "basicauth" {
		err = h.BasicAuthTokenAcquisition(apiHandler, httpClient, clientCredentials.Username, clientCredentials.Password)
	} else if h.AuthMethod == "oauth2" {
		err = h.OAuth2TokenAcquisition(apiHandler, httpClient, clientCredentials.ClientID, clientCredentials.ClientSecret)
	} else {
		err = fmt.Errorf("no valid credentials provided. Unable to obtain a token")
		h.Logger.Error("Authentication method not supported", zap.String("AuthMethod", h.AuthMethod))
	}

	if err != nil {
		h.Logger.Error("Failed to obtain new token", zap.Error(err))
	}
	return err
}

// refreshTokenIfNeeded refreshes the token if it's close to expiration.
// This function decides on the method based on the credentials type available.
func (h *AuthTokenHandler) refreshTokenIfNeeded(apiHandler apihandler.APIHandler, httpClient *http.Client, clientCredentials ClientCredentials, tokenRefreshBufferPeriod time.Duration) error {
	if time.Until(h.Expires) < tokenRefreshBufferPeriod {
		h.Logger.Info("Token is close to expiry and will be refreshed", zap.Duration("TimeUntilExpiry", time.Until(h.Expires)))
		var err error
		if clientCredentials.Username != "" && clientCredentials.Password != "" {
			err = h.RefreshBearerToken(apiHandler, httpClient)
		} else if clientCredentials.ClientID != "" && clientCredentials.ClientSecret != "" {
			err = h.OAuth2TokenAcquisition(apiHandler, httpClient, clientCredentials.ClientID, clientCredentials.ClientSecret)
		} else {
			err = fmt.Errorf("unknown auth method")
			h.Logger.Error("Failed to determine authentication method for token refresh", zap.String("AuthMethod", h.AuthMethod))
		}

		if err != nil {
			h.Logger.Error("Failed to refresh token", zap.Error(err))
			return err
		}
	}
	return nil
}
