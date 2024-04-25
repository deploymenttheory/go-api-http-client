// authenticationhandler/auth_bearer_token.go
/* The http_client_auth package focuses on authentication mechanisms for an HTTP client.
It provides structures and methods for handling both basic and bearer token based authentication */

package authenticationhandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/apiintegrations/apihandler"
	"go.uber.org/zap"
)

// BasicAuthTokenAcquisition fetches and sets an authentication token using the stored basic authentication credentials.
func (h *AuthTokenHandler) BasicAuthTokenAcquisition(apiHandler apihandler.APIHandler, httpClient *http.Client, username string, password string) error {

	// Use the APIHandler's method to get the bearer token endpoint
	bearerTokenEndpoint := apiHandler.GetBearerTokenEndpoint()

	// Construct the full authentication endpoint URL
	authenticationEndpoint := apiHandler.ConstructAPIAuthEndpoint(bearerTokenEndpoint, h.Logger)

	h.Logger.Debug("Attempting to obtain token for user", zap.String("Username", username))

	req, err := http.NewRequest("POST", authenticationEndpoint, nil)
	if err != nil {
		h.Logger.LogError("authentication_request_creation_error", "POST", authenticationEndpoint, 0, "", err, "Failed to create new request for token")
		return err
	}
	req.SetBasicAuth(username, password)

	resp, err := httpClient.Do(req)
	if err != nil {
		h.Logger.LogError("authentication_request_error", "POST", authenticationEndpoint, resp.StatusCode, resp.Status, err, "Failed to make request for token")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.Logger.LogError("token_authentication_failed", "POST", authenticationEndpoint, resp.StatusCode, resp.Status, fmt.Errorf("authentication failed with status code: %d", resp.StatusCode), "Token acquisition attempt resulted in a non-OK response")
		return fmt.Errorf("received non-OK response status: %d", resp.StatusCode)
	}

	tokenResp := &TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResp)
	if err != nil {
		h.Logger.Error("Failed to decode token response", zap.Error(err))
		return err
	}

	h.Token = tokenResp.Token
	h.Expires = tokenResp.Expires
	tokenDuration := time.Until(h.Expires)

	h.Logger.Info("Token obtained successfully", zap.Time("Expiry", h.Expires), zap.Duration("Duration", tokenDuration))

	return nil
}

// RefreshBearerToken refreshes the current authentication token.
func (h *AuthTokenHandler) RefreshBearerToken(apiHandler apihandler.APIHandler, httpClient *http.Client) error {
	h.tokenLock.Lock()
	defer h.tokenLock.Unlock()

	// Use the APIHandler's method to get the token refresh endpoint
	apiTokenRefreshEndpoint := apiHandler.GetTokenRefreshEndpoint()

	// Construct the full authentication endpoint URL
	tokenRefreshEndpoint := apiHandler.ConstructAPIAuthEndpoint(apiTokenRefreshEndpoint, h.Logger)

	h.Logger.Debug("Attempting to refresh token", zap.String("URL", tokenRefreshEndpoint))

	req, err := http.NewRequest("POST", tokenRefreshEndpoint, nil)
	if err != nil {
		h.Logger.Error("Failed to create new request for token refresh", zap.Error(err))
		return err
	}
	req.Header.Add("Authorization", "Bearer "+h.Token)

	resp, err := httpClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to make request for token refresh", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.Logger.Warn("Token refresh response status is not OK", zap.Int("StatusCode", resp.StatusCode))
		return fmt.Errorf("token refresh failed with status code: %d", resp.StatusCode)
	}

	tokenResp := &TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResp)
	if err != nil {
		h.Logger.Error("Failed to decode token response", zap.Error(err))
		return err
	}

	h.Token = tokenResp.Token
	h.Expires = tokenResp.Expires
	h.Logger.Info("Token refreshed successfully", zap.Time("Expiry", tokenResp.Expires))

	return nil
}
