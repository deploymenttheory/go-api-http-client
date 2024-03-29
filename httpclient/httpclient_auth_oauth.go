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

// ObtainOAuthToken fetches an OAuth access token using the provided OAuthCredentials (Client ID and Client Secret).
// It updates the client's Token and Expiry fields with the obtained values.
func (c *Client) ObtainOAuthToken(credentials AuthConfig) error {
	log := c.Logger

	// Use the APIHandler's method to get the OAuth token endpoint
	oauthTokenEndpoint := c.APIHandler.GetOAuthTokenEndpoint()

	// Construct the full authentication endpoint URL
	authenticationEndpoint := c.APIHandler.ConstructAPIAuthEndpoint(c.InstanceName, oauthTokenEndpoint, c.Logger)

	data := url.Values{}
	data.Set("client_id", credentials.ClientID)
	data.Set("client_secret", credentials.ClientSecret)
	data.Set("grant_type", "client_credentials")

	log.Debug("Attempting to obtain OAuth token", zap.String("ClientID", credentials.ClientID))

	req, err := http.NewRequest("POST", authenticationEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		log.Error("Failed to create request for OAuth token", zap.Error(err))
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to execute request for OAuth token", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", zap.Error(err))
		return err
	}

	// Reset the response body to its original state
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	oauthResp := &OAuthResponse{}
	err = json.Unmarshal(bodyBytes, oauthResp)
	if err != nil {
		log.Error("Failed to decode OAuth response", zap.Error(err))
		return err
	}

	if oauthResp.Error != "" {
		log.Error("Error obtaining OAuth token", zap.String("Error", oauthResp.Error))
		return fmt.Errorf("error obtaining OAuth token: %s", oauthResp.Error)
	}

	if oauthResp.AccessToken == "" {
		log.Error("Empty access token received")
		return fmt.Errorf("empty access token received")
	}

	expiresIn := time.Duration(oauthResp.ExpiresIn) * time.Second
	expirationTime := time.Now().Add(expiresIn)

	// Modified log call using the helper function
	redactedAccessToken := RedactSensitiveHeaderData(c, "AccessToken", oauthResp.AccessToken)
	log.Info("OAuth token obtained successfully", zap.String("AccessToken", redactedAccessToken), zap.Duration("ExpiresIn", expiresIn), zap.Time("ExpirationTime", expirationTime))

	c.Token = oauthResp.AccessToken
	c.Expiry = expirationTime

	return nil
}
