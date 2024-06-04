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
	"strings"
	"time"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/deploymenttheory/go-api-http-client/cookiejar"
	"github.com/deploymenttheory/go-api-http-client/headers"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/response"
	"go.uber.org/zap"
)

// DoMultiPartRequest creates and executes a multipart/form-data HTTP request for file uploads and form fields.
func (c *Client) DoMultiPartRequest(method, endpoint string, files map[string]string, params map[string]string, out interface{}) (*http.Response, error) {
	log := c.Logger
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the context is canceled when the function returns

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
	defer c.ConcurrencyHandler.ReleaseConcurrencyPermit(requestID)

	log.Debug("Executing multipart request", zap.String("method", method), zap.String("endpoint", endpoint))

	body, contentType, err := createMultipartRequestBody(files, params, log)
	if err != nil {
		return nil, err
	}

	// Log the constructed request body for debugging
	logRequestBody(body, log)

	// Construct the full URL for the API endpoint.
	url := c.APIHandler.ConstructAPIResourceEndpoint(endpoint, log)

	// Create the HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Error("Failed to create HTTP request", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	// Apply custom cookies and headers
	cookiejar.ApplyCustomCookies(req, c.clientConfig.ClientOptions.Cookies.CustomCookies, log)
	headerHandler := headers.NewHeaderHandler(req, c.Logger, c.APIHandler, c.AuthTokenHandler)
	headerHandler.SetRequestHeaders(endpoint)
	headerHandler.LogHeaders(c.clientConfig.ClientOptions.Logging.HideSensitiveData)

	// Log headers for debugging
	logHeaders(req, log)

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

// createMultipartRequestBody creates a multipart request body with the provided files and form fields.
func createMultipartRequestBody(files map[string]string, params map[string]string, log logger.Logger) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add files to the request
	for fieldName, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			log.Error("Failed to open file", zap.String("filePath", filePath), zap.Error(err))
			return nil, "", err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fieldName, filePath)
		if err != nil {
			log.Error("Failed to create form file", zap.String("fieldName", fieldName), zap.Error(err))
			return nil, "", err
		}

		fileSize, err := file.Stat()
		if err != nil {
			log.Error("Failed to get file info", zap.String("filePath", filePath), zap.Error(err))
			return nil, "", err
		}

		err = trackUploadProgress(file, part, fileSize.Size(), log)
		if err != nil {
			log.Error("Failed to copy file content", zap.String("filePath", filePath), zap.Error(err))
			return nil, "", err
		}
	}

	// Add additional parameters to the request
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err := writer.Close()
	if err != nil {
		log.Error("Failed to close writer", zap.Error(err))
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
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

// logRequestBody logs the constructed request body for debugging purposes.
func logRequestBody(body *bytes.Buffer, log logger.Logger) {
	bodyBytes := body.Bytes()
	bodyStr := string(bodyBytes)
	firstBoundaryIndex := strings.Index(bodyStr, "---")
	lastBoundaryIndex := strings.LastIndex(bodyStr, "---")
	boundaryLength := 3 // Length of "---"
	trailingCharCount := 20

	if firstBoundaryIndex != -1 && lastBoundaryIndex != -1 && firstBoundaryIndex != lastBoundaryIndex {
		// Content before the first boundary
		preBoundary := bodyStr[:firstBoundaryIndex+boundaryLength]
		// 20 characters after the first boundary
		afterFirstBoundary := bodyStr[firstBoundaryIndex+boundaryLength : firstBoundaryIndex+boundaryLength+trailingCharCount]
		// 20 characters before the last boundary
		beforeLastBoundary := bodyStr[lastBoundaryIndex-trailingCharCount : lastBoundaryIndex]
		// Everything after the last boundary
		postBoundary := bodyStr[lastBoundaryIndex:]

		log.Info("Request body preview",
			zap.String("pre_boundary", preBoundary),
			zap.String("after_first_boundary", afterFirstBoundary),
			zap.String("before_last_boundary", beforeLastBoundary),
			zap.String("post_boundary", postBoundary))
	}
}

// logHeaders logs the request headers for debugging purposes.
func logHeaders(req *http.Request, log logger.Logger) {
	for key, values := range req.Header {
		for _, value := range values {
			log.Info("Request header", zap.String("key", key), zap.String("value", value))
		}
	}
}
