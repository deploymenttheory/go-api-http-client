// http_client_oauth.go
/* The http_client_auth package focuses on authentication mechanisms for an HTTP client.
It provides structures and methods for handling OAuth-based authentication
*/
package http_client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const Authority = "https://login.microsoftonline.com/"
const Scope = "https://graph.microsoft.com/.default"

// OAuthResponse represents the response structure when obtaining an OAuth access token.
type OAuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error,omitempty"`
}

// ObtainOauthTokenWithApp fetches an OAuth access token using client credentials.
func (c *Client) ObtainOauthTokenWithApp(tenantID, clientID, clientSecret string) (*OAuthResponse, error) {
	endpoint := fmt.Sprintf("%s%s/oauth2/v2.0/token", Authority, tenantID)

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", Scope)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Debug: Print the entire raw response body for inspection
	c.logger.Debug("Raw response body: %s\n", string(bodyBytes))

	// Create a new reader from the body bytes for json unmarshalling
	bodyReader := bytes.NewReader(bodyBytes)

	oauthResp := &OAuthResponse{}
	err = json.NewDecoder(bodyReader).Decode(oauthResp)
	if err != nil {
		return nil, err
	}

	if oauthResp.Error != "" {
		return nil, fmt.Errorf("error obtaining OAuth token: %s", oauthResp.Error)
	}

	// Calculate and format token expiration time
	expiresIn := time.Duration(oauthResp.ExpiresIn) * time.Second
	expirationTime := time.Now().Add(expiresIn)
	formattedExpirationTime := expirationTime.Format(time.RFC1123)

	// Log the token life expiry details in a human-readable format
	c.logger.Debug("The OAuth token obtained is: ",
		"Valid for", expiresIn.String(),
		"Expires at", formattedExpirationTime)

	return oauthResp, nil
}

// ObtainOauthTokenWithCertificate fetches an OAuth access token using a certificate.
func (c *Client) ObtainOauthTokenWithCertificate(tenantID, clientID, thumbprint, keyFile string) (*OAuthResponse, error) {
	endpoint := fmt.Sprintf("%s%s/oauth2/v2.0/token", Authority, tenantID)

	// Load the certificate
	certData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %v", err)
	}

	cert, err := tls.X509KeyPair(certData, certData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Create a custom HTTP client with the certificate
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certData)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            certPool,
		InsecureSkipVerify: true, // Depending on your requirements, you might want to adjust this
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Prepare request data
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", Scope)
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", thumbprint) // You might need to adjust this according to your requirements
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Debug: Print the entire raw response body for inspection
	c.logger.Debug("Raw response body: %s\n", string(bodyBytes))

	bodyReader := bytes.NewReader(bodyBytes)
	oauthResp := &OAuthResponse{}
	err = json.NewDecoder(bodyReader).Decode(oauthResp)
	if err != nil {
		return nil, err
	}

	if oauthResp.Error != "" {
		return nil, fmt.Errorf("error obtaining OAuth token: %s", oauthResp.Error)
	}

	expiresIn := time.Duration(oauthResp.ExpiresIn) * time.Second
	expirationTime := time.Now().Add(expiresIn)
	formattedExpirationTime := expirationTime.Format(time.RFC1123)
	c.logger.Debug("The OAuth token obtained is: ",
		"Valid for", expiresIn.String(),
		"Expires at", formattedExpirationTime)

	return oauthResp, nil
}

// GetOAuthCredentials retrieves the current OAuth credentials (Client ID and Client Secret)
// set for the client instance. Used for test cases.
func (c *Client) GetOAuthCredentials() OAuthCredentials {
	return c.OAuthCredentials
}
