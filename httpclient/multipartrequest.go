// httpclient/multipartrequest.go
package httpclient

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/deploymenttheory/go-api-http-client/headers"
	"github.com/deploymenttheory/go-api-http-client/helpers"
	"github.com/deploymenttheory/go-api-http-client/response"
	"go.uber.org/zap"
)

// DoMultipartRequest creates and executes a multipart HTTP request. It is used for sending files
// and form fields in a single request. This method handles the construction of the multipart
// message body, setting the appropriate headers, and sending the request to the given endpoint.
//
// Parameters:
// - method: The HTTP method to use (e.g., POST, PUT).
// - endpoint: The API endpoint to which the request will be sent.
// - fields: A map of form fields and their values to include in the multipart message.
// - files: A map of file field names to file paths that will be included as file attachments.
// - out: A pointer to a variable where the unmarshaled response will be stored.
//
// Returns:
// - A pointer to the http.Response received from the server.
// - An error if the request could not be sent or the response could not be processed.
//
// The function first validates the authentication token, then constructs the multipart
// request body based on the provided fields and files. It then constructs the full URL for
// the request, sets the required headers (including Authorization and Content-Type), and
// sends the request.
//
// If debug mode is enabled, the function logs all the request headers before sending the request.
// After the request is sent, the function checks the response status code. If the response is
// not within the success range (200-299), it logs an error and returns the response and an error.
// If the response is successful, it attempts to unmarshal the response body into the 'out' parameter.
//
// Note:
// The caller should handle closing the response body when successful.
func (c *Client) DoMultipartRequest(method, endpoint string, fields map[string]string, files map[string]string, out interface{}) (*http.Response, error) {
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

	// Create a buffer to hold the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the simple fields to the form data
	for field, value := range fields {
		log.Debug("Adding field to multipart request", zap.String("Field", field), zap.String("Value", value))
		if err := writer.WriteField(field, value); err != nil {
			return nil, err
		}
	}

	// Add the files to the form data
	for formField, filePath := range files {
		file, err := helpers.SafeOpenFile(filePath)
		if err != nil {
			log.Error("Failed to open file securely", zap.String("file", filePath), zap.Error(err))
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(formField, filepath.Base(filePath))
		if err != nil {
			return nil, err
		}
		log.Debug("Adding file to multipart request", zap.String("FormField", formField), zap.String("FilePath", filePath))
		if _, err := io.Copy(part, file); err != nil {
			return nil, err
		}
	}

	// Close the writer to finish writing the multipart message
	if err := writer.Close(); err != nil {
		return nil, err
	}

	contentType := writer.FormDataContentType()
	bodyBytes := body.Bytes()

	// Extract the first and last parts of the body for logging
	const logSegmentSize = 1024 // 1 KB
	bodyLen := len(bodyBytes)
	var logBody string
	if bodyLen <= 2*logSegmentSize {
		logBody = string(bodyBytes)
	} else {
		logBody = string(bodyBytes[:logSegmentSize]) + "..." + string(bodyBytes[bodyLen-logSegmentSize:])
	}

	// Log the boundary and a partial body for debugging
	boundary := writer.Boundary()
	log.Debug("Multipart boundary", zap.String("Boundary", boundary))
	log.Debug("Multipart request body (partial)", zap.String("Body", logBody))

	// Construct URL using the ConstructAPIResourceEndpoint function
	url := c.APIHandler.ConstructAPIResourceEndpoint(endpoint, log)

	// Create the request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	// Initialize HeaderManager
	headerHandler := headers.NewHeaderHandler(req, c.Logger, c.APIHandler, c.AuthTokenHandler)

	// Use HeaderManager to set headers
	headerHandler.SetContentType(contentType)
	headerHandler.SetRequestHeaders(endpoint)
	headerHandler.LogHeaders(c.clientConfig.ClientOptions.Logging.HideSensitiveData)

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to send multipart request", zap.String("method", method), zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}

	// Check for successful status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Handle error responses
		return nil, response.HandleAPIErrorResponse(resp, log)
	} else {
		// Handle successful responses
		return resp, response.HandleAPISuccessResponse(resp, out, log)
	}
}
