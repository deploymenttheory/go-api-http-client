// http_client_bearer_token_auth.go
/* The http_client_auth package focuses on authentication mechanisms for an HTTP client.
It provides structures and methods for handling both basic and bearer token based authentication
*/
package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// BearerTokenAuthCredentials represents the username and password for basic authentication.
type BearerTokenAuthCredentials struct {
	Username string
	Password string
}

// SetBearerTokenAuthCredentials sets the BearerTokenAuthCredentials (Username and Password)
// for the client instance. These credentials are used for obtaining and refreshing
// bearer tokens for authentication.
func (c *Client) SetBearerTokenAuthCredentials(credentials BearerTokenAuthCredentials) {
	c.BearerTokenAuthCredentials = credentials
}

// ObtainToken fetches and sets an authentication token using the stored basic authentication credentials.
func (c *Client) ObtainToken(log logger.Logger) error {

	// Use the APIHandler's method to get the bearer token endpoint
	bearerTokenEndpoint := c.APIHandler.GetBearerTokenEndpoint()

	// Construct the full authentication endpoint URL
	authenticationEndpoint := c.APIHandler.ConstructAPIAuthEndpoint(c.InstanceName, bearerTokenEndpoint, c.Logger)

	log.Debug("Attempting to obtain token for user", zap.String("Username", c.BearerTokenAuthCredentials.Username))

	req, err := http.NewRequest("POST", authenticationEndpoint, nil)
	if err != nil {
		log.LogError("authentication_request_creation_error", "POST", authenticationEndpoint, 0, err, "Failed to create new request for token")
		return err
	}
	req.SetBasicAuth(c.BearerTokenAuthCredentials.Username, c.BearerTokenAuthCredentials.Password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.LogError("authentication_request_error", "POST", authenticationEndpoint, 0, err, "Failed to make request for token")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.LogError("token_authentication_failed", "POST", authenticationEndpoint, resp.StatusCode, fmt.Errorf("authentication failed with status code: %d", resp.StatusCode), "Token acquisition attempt resulted in a non-OK response")
		return fmt.Errorf("received non-OK response status: %d", resp.StatusCode)
	}

	tokenResp := &TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResp)
	if err != nil {
		log.Error("Failed to decode token response", zap.Error(err))
		return err
	}

	c.Token = tokenResp.Token
	c.Expiry = tokenResp.Expires
	tokenDuration := time.Until(c.Expiry)

	log.Info("Token obtained successfully", zap.Time("Expiry", c.Expiry), zap.Duration("Duration", tokenDuration))

	return nil
}

// RefreshToken refreshes the current authentication token.
func (c *Client) RefreshToken(log logger.Logger) error {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	apiTokenRefreshEndpoint := c.APIHandler.GetTokenRefreshEndpoint()

	// Construct the full authentication endpoint URL
	tokenRefreshEndpoint := c.APIHandler.ConstructAPIAuthEndpoint(c.InstanceName, apiTokenRefreshEndpoint, c.Logger)

	req, err := http.NewRequest("POST", tokenRefreshEndpoint, nil)
	if err != nil {
		log.Error("Failed to create new request for token refresh", zap.Error(err))
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)

	log.Debug("Attempting to refresh token", zap.String("URL", tokenRefreshEndpoint))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to make request for token refresh", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("Token refresh response status is not OK", zap.Int("StatusCode", resp.StatusCode))
		return err
	}

	tokenResp := &TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResp)
	if err != nil {
		log.Error("Failed to decode token response", zap.Error(err))
		return err
	}

	c.Token = tokenResp.Token
	c.Expiry = tokenResp.Expires
	log.Info("Token refreshed successfully", zap.Time("Expiry", tokenResp.Expires))

	return nil
}
