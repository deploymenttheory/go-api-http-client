// httpclient/download.go
package httpclient

import (
	"io"
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/deploymenttheory/go-api-http-client/headers"
	"github.com/deploymenttheory/go-api-http-client/response"
	"go.uber.org/zap"
)

// DoDownloadRequest performs a download from a given URL. It follows the same authentication,
// header setting, and URL construction as the DoMultipartRequest function. The downloaded data
// is written to the provided writer.
//
// Parameters:
// - method: The HTTP method to use (e.g., GET).
// - endpoint: The API endpoint from which the file will be downloaded.
// - out: A writer where the downloaded data will be written.
//
// Returns:
// - A pointer to the http.Response received from the server.
// - An error if the request could not be sent or the response could not be processed.
//
// The function first validates the authentication token, constructs the full URL for
// the request, sets the required headers (including Authorization), and sends the request.
//
// If debug mode is enabled, the function logs all the request headers before sending the request.
// After the request is sent, the function checks the response status code. If the response is
// not within the success range (200-299), it logs an error and returns the response and an error.
// If the response is successful, the function writes the response body to the provided writer.
//
// Note:
// The caller should handle closing the response body when successful.
func (c *Client) DoDownloadRequest(method, endpoint string, out io.Writer) (*http.Response, error) {
	log := c.Logger

	// Auth Token validation check
	clientCredentials := authenticationhandler.ClientCredentials{
		Username:     c.clientConfig.Auth.Username,
		Password:     c.clientConfig.Auth.Password,
		ClientID:     c.clientConfig.Auth.ClientID,
		ClientSecret: c.clientConfig.Auth.ClientSecret,
	}

	valid, err := c.AuthTokenHandler.CheckAndRefreshAuthToken(c.APIHandler, c.httpClient, clientCredentials, c.clientConfig.ClientOptions.Timeout.TokenRefreshBufferPeriod.Duration())
	if err != nil || !valid {
		return nil, err
	}

	// Construct URL using the ConstructAPIResourceEndpoint function
	url := c.APIHandler.ConstructAPIResourceEndpoint(endpoint, log)

	// Create the request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Initialize HeaderManager
	headerHandler := headers.NewHeaderHandler(req, c.Logger, c.APIHandler, c.AuthTokenHandler)

	// Use HeaderManager to set headers
	headerHandler.SetRequestHeaders(endpoint)
	headerHandler.LogHeaders(c.clientConfig.ClientOptions.Logging.HideSensitiveData)

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to send download request", zap.String("method", method), zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}

	// Check for successful status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Handle error responses
		return nil, response.HandleAPIErrorResponse(resp, log)
	}

	// Write the response body to the provided writer
	defer resp.Body.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return nil, err
	}

	return resp, nil
}
