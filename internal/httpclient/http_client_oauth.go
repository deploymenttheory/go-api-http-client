// http_client_oauth.go
/* The http_client_auth package focuses on authentication mechanisms for an HTTP client.
It provides structures and methods for handling OAuth-based authentication
*/
package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// OAuthResponse represents the response structure when obtaining an OAuth access token.
type OAuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error,omitempty"`
}

// OAuthCredentials contains the client ID and client secret required for OAuth authentication.
type OAuthCredentials struct {
	ClientID     string
	ClientSecret string
}

// SetOAuthCredentials sets the OAuth credentials (Client ID and Client Secret)
// for the client instance. These credentials are used for obtaining and refreshing
// OAuth tokens for authentication.
func (c *Client) SetOAuthCredentials(credentials OAuthCredentials) {
	c.OAuthCredentials = credentials
}

// ObtainOAuthToken fetches an OAuth access token using the provided OAuthCredentials (Client ID and Client Secret).
// It updates the client's Token and Expiry fields with the obtained values.
func (c *Client) ObtainOAuthToken(credentials AuthConfig) error {
	authenticationEndpoint := c.ConstructAPIAuthEndpoint(OAuthTokenEndpoint)
	data := url.Values{}
	data.Set("client_id", credentials.ClientID)
	data.Set("client_secret", credentials.ClientSecret)
	data.Set("grant_type", "client_credentials")

	c.logger.Debug("Attempting to obtain OAuth token", zap.String("ClientID", credentials.ClientID))

	req, err := http.NewRequest("POST", authenticationEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		c.logger.Error("Failed to create request for OAuth token", zap.Error(err))
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to execute request for OAuth token", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body", zap.Error(err))
		return err
	}

	// Reset the response body to its original state
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	oauthResp := &OAuthResponse{}
	err = json.Unmarshal(bodyBytes, oauthResp)
	if err != nil {
		c.logger.Error("Failed to decode OAuth response", zap.Error(err))
		return err
	}

	if oauthResp.Error != "" {
		c.logger.Error("Error obtaining OAuth token", zap.String("Error", oauthResp.Error))
		return fmt.Errorf("error obtaining OAuth token: %s", oauthResp.Error)
	}

	if oauthResp.AccessToken == "" {
		c.logger.Error("Empty access token received")
		return fmt.Errorf("empty access token received")
	}

	expiresIn := time.Duration(oauthResp.ExpiresIn) * time.Second
	expirationTime := time.Now().Add(expiresIn)
	c.logger.Info("OAuth token obtained successfully", zap.String("AccessToken", oauthResp.AccessToken), zap.Duration("ExpiresIn", expiresIn), zap.Time("ExpirationTime", expirationTime))

	c.Token = oauthResp.AccessToken
	c.Expiry = expirationTime

	return nil
}

// InvalidateOAuthToken invalidates the current OAuth access token.
// After invalidation, the token cannot be used for further API requests.
func (c *Client) InvalidateOAuthToken() error {
	invalidateTokenEndpoint := c.ConstructAPIAuthEndpoint(TokenInvalidateEndpoint)

	c.logger.Debug("Attempting to invalidate OAuth token", zap.String("Endpoint", invalidateTokenEndpoint))

	req, err := http.NewRequest("POST", invalidateTokenEndpoint, nil)
	if err != nil {
		c.logger.Error("Failed to create new request for token invalidation", zap.Error(err))
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to make request for token invalidation", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errMsg := fmt.Errorf("failed to invalidate token, status code: %d", resp.StatusCode)
		c.logger.Error("Failed to invalidate OAuth token", zap.Int("StatusCode", resp.StatusCode), zap.Error(errMsg))
		return errMsg
	}

	c.logger.Info("OAuth token invalidated successfully", zap.String("Endpoint", invalidateTokenEndpoint))

	return nil
}
