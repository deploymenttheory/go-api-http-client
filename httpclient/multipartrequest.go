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

	for fieldName, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			log.Error("Failed to open file", zap.String("filePath", filePath), zap.Error(err))
			return nil, "", err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(fieldName, file.Name())
		if err != nil {
			log.Error("Failed to create form file", zap.String("fieldName", fieldName), zap.Error(err))
			return nil, "", err
		}

		fileSize, err := file.Stat()
		if err != nil {
			log.Error("Failed to get file info", zap.String("filePath", filePath), zap.Error(err))
			return nil, "", err
		}

		// Start logging the progress
		progressLogger := logUploadProgress(fileSize.Size(), log)

		// Chunk the file upload and log the progress
		err = chunkFileUpload(file, part, log, progressLogger)
		if err != nil {
			log.Error("Failed to copy file content", zap.String("filePath", filePath), zap.Error(err))
			return nil, "", err
		}
	}

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

// chunkFileUpload reads the file in chunks and writes it to the writer.
func chunkFileUpload(file *os.File, writer io.Writer, log logger.Logger, updateProgress func(int64)) error {
	const logThreshold = 1 << 20 // 1 MB in bytes
	buffer := make([]byte, 4096)
	totalWritten := int64(0)
	chunkWritten := int64(0)

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		written, err := writer.Write(buffer[:n])
		if err != nil {
			return err
		}

		totalWritten += int64(written)
		chunkWritten += int64(written)
		updateProgress(int64(written))

		// Log progress for every 1MB chunk written
		if chunkWritten >= logThreshold {
			log.Debug("Chunk written", zap.Int64("bytes_written", chunkWritten), zap.Int64("total_written", totalWritten))
			chunkWritten = 0
		}
	}

	// Log any remaining bytes that were written but didn't reach the log threshold
	if chunkWritten > 0 {
		log.Debug("Chunk written", zap.Int64("bytes_written", chunkWritten), zap.Int64("total_written", totalWritten))
	}

	return nil
}

// trackUploadProgress logs the upload progress based on the percentage of the total upload.
func logUploadProgress(totalSize int64, log logger.Logger) func(int64) {
	var uploadedSize int64
	var lastLoggedPercentage float64
	startTime := time.Now()

	return func(bytesWritten int64) {
		uploadedSize += bytesWritten
		percentage := math.Floor(float64(uploadedSize) / float64(totalSize) * 100)
		uploadedMB := float64(uploadedSize) / (1024 * 1024)

		if percentage != lastLoggedPercentage {
			log.Debug("File upload progress",
				zap.String("completed", fmt.Sprintf("%.0f%%", percentage)),
				zap.Float64("uploaded_megabytes", uploadedMB),
				zap.Duration("elapsed_time", time.Since(startTime)))
			lastLoggedPercentage = percentage
		}

		if uploadedSize == totalSize {
			totalTime := time.Since(startTime)
			log.Info("File upload completed",
				zap.Float64("total_uploaded_megabytes", float64(uploadedSize)/(1024*1024)),
				zap.Duration("total_upload_time", totalTime))
		}
	}
}

// logRequestBody logs the constructed request body for debugging purposes.
func logRequestBody(body *bytes.Buffer, log logger.Logger) {
	bodyBytes := body.Bytes()
	bodyStr := string(bodyBytes)

	// Find the boundary string
	boundaryIndex := strings.Index(bodyStr, "\r\n")
	if boundaryIndex == -1 {
		log.Warn("No boundary found in request body")
		return
	}
	boundary := bodyStr[:boundaryIndex]

	// Split the body by boundaries
	parts := strings.Split(bodyStr, boundary)

	var loggedParts []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "--" || part == "" {
			continue // Skip the last boundary marker or empty parts
		}
		if strings.Contains(part, "Content-Disposition: form-data; name=\"file\"") {
			// If it's a file part, only log the headers
			headersEndIndex := strings.Index(part, "\r\n\r\n")
			if headersEndIndex != -1 {
				headers := part[:headersEndIndex]
				loggedParts = append(loggedParts, headers+"\r\n\r\n<file content omitted>\r\n")
			} else {
				loggedParts = append(loggedParts, part)
			}
		} else {
			// Otherwise, log the entire part
			loggedParts = append(loggedParts, part)
		}
	}

	// Join the logged parts back together with the boundary
	loggedBody := boundary + "\r\n" + strings.Join(loggedParts, "\r\n"+boundary+"\r\n") + "\r\n" + boundary + "--"

	log.Info("Request body preview", zap.String("body", loggedBody))
}

// logHeaders logs the request headers for debugging purposes.
func logHeaders(req *http.Request, log logger.Logger) {
	for key, values := range req.Header {
		for _, value := range values {
			log.Info("Request header", zap.String("key", key), zap.String("value", value))
		}
	}
}
