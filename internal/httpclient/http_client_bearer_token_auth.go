// http_client_bearer_token_auth.go
/* The http_client_auth package focuses on authentication mechanisms for an HTTP client.
It provides structures and methods for handling both basic and bearer token based authentication
*/
package httpclient

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-api-http-client/internal/errors"
	"github.com/deploymenttheory/go-api-http-client/internal/logger"
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

	bearerTokenEndpoint := c.APIHandler.GetBearerTokenEndpoint()
	authenticationEndpoint := c.APIHandler.ConstructAPIAuthEndpoint(bearerTokenEndpoint, c.Logger)

	log.Debug("Attempting to obtain token for user", zap.String("Username", c.BearerTokenAuthCredentials.Username))

	req, err := http.NewRequest("POST", authenticationEndpoint, nil)
	if err != nil {
		log.Error("Failed to create new request for token", zap.Error(err))
		return err
	}
	req.SetBasicAuth(c.BearerTokenAuthCredentials.Username, c.BearerTokenAuthCredentials.Password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to make request for token", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("Received non-OK response while obtaining token", zap.Int("StatusCode", resp.StatusCode))
		return errors.HandleAPIError(resp, log)
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

/*
// RefreshToken refreshes the current authentication token.
func (c *Client) RefreshToken() error {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	tokenRefreshEndpoint := c.ConstructAPIAuthEndpoint(TokenRefreshEndpoint)

	req, err := http.NewRequest("POST", tokenRefreshEndpoint, nil)
	if err != nil {
		log.Error("Failed to create new request for token refresh", "error", err)
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)

	log.Debug("Attempting to refresh token", "URL", tokenRefreshEndpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to make request for token refresh", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn("Token refresh response status is not OK", "StatusCode", resp.StatusCode)
		return c.HandleAPIError(resp)
	}

	tokenResp := &TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResp)
	if err != nil {
		log.Error("Failed to decode token response", "error", err)
		return err
	}

	log.Info("Token refreshed successfully", "Expiry", tokenResp.Expires)

	c.Token = tokenResp.Token
	c.Expiry = tokenResp.Expires
	return nil
}
*/
// RefreshToken refreshes the current authentication token.
func (c *Client) RefreshToken(log logger.Logger) error {
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()

	apiTokenRefreshEndpoint := c.APIHandler.GetTokenRefreshEndpoint()

	tokenRefreshEndpoint := c.APIHandler.ConstructAPIAuthEndpoint(apiTokenRefreshEndpoint, c.Logger)

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
		return errors.HandleAPIError(resp, log)
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
