// authenticationhandler/auth_oauth.go

/* The http_client_auth package focuses on authentication mechanisms for an HTTP client.
It provides structures and methods for handling OAuth-based authentication
*/

package authenticationhandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/deploymenttheory/go-api-http-client/apiintegrations/apihandler"
	"github.com/deploymenttheory/go-api-http-client/headers/redact"
	"go.uber.org/zap"
)

// OAuthResponse represents the response structure when obtaining an OAuth access token.
type OAuthResponse struct {
	AccessToken  string `json:"access_token"`            // AccessToken is the token that can be used in subsequent requests for authentication.
	ExpiresIn    int64  `json:"expires_in"`              // ExpiresIn specifies the duration in seconds after which the access token expires.
	TokenType    string `json:"token_type"`              // TokenType indicates the type of token, typically "Bearer".
	RefreshToken string `json:"refresh_token,omitempty"` // RefreshToken is used to obtain a new access token when the current one expires.
	Error        string `json:"error,omitempty"`         // Error contains details if an error occurs during the token acquisition process.
}

// ObtainOAuthToken fetches an OAuth access token using the provided client ID and client secret.
// It updates the AuthTokenHandler's Token and Expires fields with the obtained values.
func (h *AuthTokenHandler) ObtainOAuthToken(apiHandler apihandler.APIHandler, httpClient *http.Client, clientID, clientSecret string) error {
	// Get the OAuth token endpoint from the APIHandler
	oauthTokenEndpoint := apiHandler.GetOAuthTokenEndpoint()

	// Construct the full authentication endpoint URL
	authenticationEndpoint := apiHandler.ConstructAPIAuthEndpoint(oauthTokenEndpoint, h.Logger)

	// Get the OAuth token scope from the APIHandler
	oauthTokenScope := apiHandler.GetOAuthTokenScope()

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", oauthTokenScope)
	data.Set("grant_type", "client_credentials")

	h.Logger.Debug("Attempting to obtain OAuth token", zap.String("ClientID", clientID), zap.String("Scope", oauthTokenScope))

	req, err := http.NewRequest("POST", authenticationEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		h.Logger.Error("Failed to create request for OAuth token", zap.Error(err))
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		h.Logger.Error("Failed to execute request for OAuth token", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		h.Logger.Error("Failed to read response body", zap.Error(err))
		return err
	}

	// Reset the response body to its original state
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	oauthResp := &OAuthResponse{}
	err = json.Unmarshal(bodyBytes, oauthResp)
	if err != nil {
		h.Logger.Error("Failed to decode OAuth response", zap.Error(err))
		return err
	}

	if oauthResp.Error != "" {
		h.Logger.Error("Error obtaining OAuth token", zap.String("Error", oauthResp.Error))
		return fmt.Errorf("error obtaining OAuth token: %s", oauthResp.Error)
	}

	if oauthResp.AccessToken == "" {
		h.Logger.Error("Empty access token received")
		return fmt.Errorf("empty access token received")
	}

	expiresIn := time.Duration(oauthResp.ExpiresIn) * time.Second
	expirationTime := time.Now().Add(expiresIn)

	// Modified log call using the helper function
	redactedAccessToken := redact.RedactSensitiveHeaderData(h.HideSensitiveData, "AccessToken", oauthResp.AccessToken)
	h.Logger.Info("OAuth token obtained successfully", zap.String("AccessToken", redactedAccessToken), zap.Duration("ExpiresIn", expiresIn), zap.Time("ExpirationTime", expirationTime))

	h.Token = oauthResp.AccessToken
	h.Expires = expirationTime

	return nil
}
