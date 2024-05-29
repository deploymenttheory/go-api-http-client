// authenticationhandler/basicauthentication.go
/* The http_client_auth package focuses on authentication mechanisms for an HTTP client.
It provides structures and methods for handling both basic and bearer token based authentication */

package httpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// TokenResponse represents the structure of a token response from the API.

func (c *Client) GetBasicToken() error {

	// NOTE this must return the correct full auth enpoint, figuring out which one to return depending on client config
	endpoint := c.API.GetBearerAuthEndpoint(c.Logger)

	c.Logger.Debug("Attempting to obtain token for user", zap.String("Username", c.config.BasicAuthUsername))

	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		c.Logger.LogError("authentication_request_creation_error", "POST", endpoint, 0, "", err, "Failed to create new request for token")
		return err
	}
	req.SetBasicAuth(c.config.BasicAuthUsername, c.config.BasicAuthPassword)

	resp, err := c.http.Do(req)
	if err != nil {
		c.Logger.LogError("authentication_request_error", "POST", endpoint, 0, "", err, "Failed to make request for token")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.Logger.LogError("token_authentication_failed", "POST", endpoint, resp.StatusCode, resp.Status, fmt.Errorf("authentication failed with status code: %d", resp.StatusCode), "Token acquisition attempt resulted in a non-OK response")
		return fmt.Errorf("received non-OK response status: %d", resp.StatusCode)
	}

	// TODO generalise this struct somehow
	tokenResp := &TokenResponse{}
	err = json.NewDecoder(resp.Body).Decode(tokenResp)
	if err != nil {
		c.Logger.Error("Failed to decode token response", zap.Error(err))
		return err
	}

	c.AuthToken = tokenResp.Token
	c.AuthTokenExpiry = tokenResp.Expires
	tokenDuration := time.Until(c.AuthTokenExpiry)

	c.Logger.Info("Token obtained successfully", zap.Time("Expiry", c.AuthTokenExpiry), zap.Duration("Duration", tokenDuration))

	return nil
}
