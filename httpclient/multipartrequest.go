// httpclient/multipart_request.go
package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/deploymenttheory/go-api-http-client/cookiejar"
	"github.com/deploymenttheory/go-api-http-client/headers"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/response"
	"go.uber.org/zap"
)

// DoMultiPartRequest creates and executes a multipart/form-data HTTP request for file uploads and form fields.
// This method handles the construction of the multipart message body, setting the appropriate headers, and
// sending the request to the specified endpoint.
//
// Parameters:
// - method: The HTTP method to use (e.g., POST, PUT). Only POST and PUT are supported for multipart requests.
// - endpoint: The API endpoint to which the request will be sent.
// - files: A map of form field names to file paths. The files will be read from the disk and included as file attachments.
// - params: A map of additional form fields and their values to include in the multipart message.
// - out: A pointer to a variable where the unmarshaled response will be stored.
//
// Returns:
// - A pointer to the http.Response received from the server.
// - An error if the request could not be sent or the response could not be processed.
//
// The function first validates the authentication token, constructs the multipart request body with the provided files and fields,
// and then constructs the full URL for the request. It sets the required headers (including Authorization and Content-Type), and sends the request.
// The function tracks and logs the file upload progress in terms of the percentage of the total upload.
//
// Debug logging is used throughout the process to log significant steps, including request creation, header settings, and upload progress.
// After sending the request, the function checks the response status code. If the response is not within the success range (200-299), it logs an error
// and returns the response and an error. If the response is successful, it attempts to unmarshal the response body into the 'out' parameter.
//
// Note:
// The caller should handle closing the response body when the response is not nil to prevent resource leaks.
func (c *Client) DoMultiPartRequest(method, endpoint string, files map[string]string, params map[string]string, out interface{}) (*http.Response, error) {
	log := c.Logger
	ctx := context.Background()

	// Ensure the method is supported
	if method != http.MethodPost && method != http.MethodPut {
		log.Error("HTTP method not supported for multipart request", zap.String("method", method))
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// Authenticate the client using the provided credentials and refresh the auth token if necessary.
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

	// Acquire a concurrency permit to control the number of concurrent requests.
	ctx, requestID, err := c.ConcurrencyHandler.AcquireConcurrencyPermit(ctx)
	if err != nil {
		log.Error("Failed to acquire concurrency permit", zap.Error(err))
		return nil, err
	}

	// Ensure the concurrency permit is released after the function exits.
	defer func() {
		c.ConcurrencyHandler.ReleaseConcurrencyPermit(requestID)
	}()

	log.Debug("Executing multipart request", zap.String("method", method), zap.String("endpoint", endpoint))

	// Create a new multipart writer to construct the request body.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add files to the request
	for fieldName, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			log.Error("Failed to open file", zap.String("filePath", filePath), zap.Error(err))
			return nil, err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fieldName, filePath)
		if err != nil {
			log.Error("Failed to create form file", zap.String("fieldName", fieldName), zap.Error(err))
			return nil, err
		}

		fileSize, err := file.Stat()
		if err != nil {
			log.Error("Failed to get file info", zap.String("filePath", filePath), zap.Error(err))
			return nil, err
		}

		err = trackUploadProgress(file, part, fileSize.Size(), log)
		if err != nil {
			log.Error("Failed to copy file content", zap.String("filePath", filePath), zap.Error(err))
			return nil, err
		}
	}

	// Add additional parameters to the request
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		log.Error("Failed to close writer", zap.Error(err))
		return nil, err
	}

	// Construct the full URL for the API endpoint.
	url := c.APIHandler.ConstructAPIResourceEndpoint(endpoint, log)

	// Create the HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Error("Failed to create HTTP request", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Apply custom cookies and headers
	cookiejar.ApplyCustomCookies(req, c.clientConfig.ClientOptions.Cookies.CustomCookies, log)
	headerHandler := headers.NewHeaderHandler(req, c.Logger, c.APIHandler, c.AuthTokenHandler)
	headerHandler.SetRequestHeaders(endpoint)
	headerHandler.LogHeaders(c.clientConfig.ClientOptions.Logging.HideSensitiveData)

	req = req.WithContext(ctx)

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error("Failed to send request", zap.String("method", method), zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}

	duration := time.Since(startTime)
	log.Debug("Request sent successfully", zap.String("method", method), zap.String("endpoint", endpoint), zap.Int("status_code", resp.StatusCode), zap.Duration("duration", duration))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, response.HandleAPISuccessResponse(resp, out, log)
	}

	return resp, response.HandleAPIErrorResponse(resp, log)
}

// trackUploadProgress logs the upload progress based on the percentage of the total upload.
func trackUploadProgress(file *os.File, writer io.Writer, totalSize int64, log logger.Logger) error {
	buffer := make([]byte, 4096)
	var uploadedSize int64
	var lastLoggedPercentage float64
	startTime := time.Now()

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		uploadedSize += int64(n)
		percentage := math.Floor(float64(uploadedSize) / float64(totalSize) * 100)
		uploadedMB := float64(uploadedSize) / (1024 * 1024)

		if percentage != lastLoggedPercentage {
			log.Debug("File upload progress",
				zap.String("completed", fmt.Sprintf("%.0f%%", percentage)),
				zap.Float64("uploaded_megabytes", uploadedMB),
				zap.Duration("elapsed_time", time.Since(startTime)))
			lastLoggedPercentage = percentage
		}

		_, err = writer.Write(buffer[:n])
		if err != nil {
			return err
		}
	}

	totalTime := time.Since(startTime)
	log.Info("File upload completed",
		zap.Float64("total_uploaded_megabytes", float64(uploadedSize)/(1024*1024)),
		zap.Duration("total_upload_time", totalTime))

	return nil
}
