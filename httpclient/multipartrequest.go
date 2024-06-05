package httpclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
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
func (c *Client) DoMultiPartRequest(method, endpoint string, files map[string][]string, params map[string]string, contentTypes map[string]string, headersMap map[string]http.Header, out interface{}) (*http.Response, error) {
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

	body, contentType, err := createMultipartRequestBody(files, params, contentTypes, headersMap, log)
	if err != nil {
		return nil, err
	}

	// Log the constructed request body for debugging
	logMultiPartRequestBody(body, log)

	// Construct the full URL for the API endpoint.
	url := c.APIHandler.ConstructAPIResourceEndpoint(endpoint, log)

	// Create the HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Error("Failed to create HTTP request", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

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

// createMultipartRequestBody creates a multipart request body with the provided files and form fields, supporting custom content types and headers.
func createMultipartRequestBody(files map[string][]string, params map[string]string, contentTypes map[string]string, headersMap map[string]http.Header, log logger.Logger) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for fieldName, filePaths := range files {
		for _, filePath := range filePaths {
			if err := addFilePart(writer, fieldName, filePath, contentTypes, headersMap, log); err != nil {
				return nil, "", err
			}
		}
	}

	for key, val := range params {
		if err := addFormField(writer, key, val, log); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		log.Error("Failed to close writer", zap.Error(err))
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

// addFilePart adds a base64 encoded file part to the multipart writer with the provided field name and file path.
func addFilePart(writer *multipart.Writer, fieldName, filePath string, contentTypes map[string]string, headersMap map[string]http.Header, log logger.Logger) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Failed to open file", zap.String("filePath", filePath), zap.Error(err))
		return err
	}
	defer file.Close()

	contentType := "application/octet-stream"
	if ct, ok := contentTypes[fieldName]; ok {
		contentType = ct
	}

	var part io.Writer
	if h, ok := headersMap[fieldName]; ok {
		part, err = writer.CreatePart(CustomFormDataHeader(fieldName, filepath.Base(filePath), contentType, h))
	} else {
		part, err = writer.CreateFormFile(fieldName, filepath.Base(filePath))
	}

	if err != nil {
		log.Error("Failed to create form file part", zap.String("fieldName", fieldName), zap.Error(err))
		return err
	}

	encoder := base64.NewEncoder(base64.StdEncoding, part)
	defer encoder.Close()

	fileSize, err := file.Stat()
	if err != nil {
		log.Error("Failed to get file info", zap.String("filePath", filePath), zap.Error(err))
		return err
	}

	progressLogger := logUploadProgress(fileSize.Size(), log)
	if err := chunkFileUpload(file, encoder, log, progressLogger); err != nil {
		log.Error("Failed to copy file content", zap.String("filePath", filePath), zap.Error(err))
		return err
	}

	return nil
}

// addFormField adds a form field to the multipart writer with the provided key and value.
func addFormField(writer *multipart.Writer, key, val string, log logger.Logger) error {
	fieldWriter, err := writer.CreateFormField(key)
	if err != nil {
		log.Error("Failed to create form field", zap.String("key", key), zap.Error(err))
		return err
	}
	if _, err := fieldWriter.Write([]byte(val)); err != nil {
		log.Error("Failed to write form field", zap.String("key", key), zap.Error(err))
		return err
	}
	return nil
}

// CustomFormDataHeader creates a textproto.MIMEHeader for a form data field with the provided field name, file name, content type, and custom headers.
func CustomFormDataHeader(fieldname, filename, contentType string, customHeaders http.Header) textproto.MIMEHeader {
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldname, filename))
	header.Set("Content-Type", contentType)
	for key, values := range customHeaders {
		for _, value := range values {
			header.Add(key, value)
		}
	}
	return header
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
		log.Debug("Final chunk written", zap.Int64("bytes_written", chunkWritten), zap.Int64("total_written", totalWritten))
	}

	return nil
}

// logUploadProgress logs the upload progress based on the percentage of the total upload.
func logUploadProgress(totalSize int64, log logger.Logger) func(int64) {
	var uploadedSize int64
	var lastLoggedPercentage float64
	startTime := time.Now()

	return func(bytesWritten int64) {
		uploadedSize += bytesWritten
		percentage := math.Floor(float64(uploadedSize) / float64(totalSize) * 100)
		uploadedMB := float64(uploadedSize) / (1024 * 1024)

		if percentage != lastLoggedPercentage {
			log.Info("File upload progress",
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

// logMultiPartRequestBody logs the constructed request body for debugging purposes.
func logMultiPartRequestBody(body *bytes.Buffer, log logger.Logger) {
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

		headersEndIndex := strings.Index(part, "\r\n\r\n")
		if headersEndIndex != -1 {
			headers := part[:headersEndIndex]
			bodyContent := part[headersEndIndex+4:]

			encoding := "none"
			if strings.Contains(headers, "base64") || strings.Contains(bodyContent, "base64,") {
				encoding = "base64"
			} else if strings.Contains(headers, "quoted-printable") {
				encoding = "quoted-printable"
			}

			// Log headers and indicate content is omitted
			if strings.Contains(headers, "Content-Disposition: form-data; name=\"file\"") {
				log.Info("Multipart section",
					zap.String("content_disposition", headers),
					zap.String("encoding", encoding))
				loggedParts = append(loggedParts, headers+"\r\n\r\n<file content omitted>")
			} else {
				log.Info("Multipart section",
					zap.String("content_disposition", headers),
					zap.String("encoding", encoding))
				// Log the entire part if it's not a file
				loggedParts = append(loggedParts, headers+"\r\n\r\n"+bodyContent)
			}
		} else {
			loggedParts = append(loggedParts, part)
		}
	}

	// Join the logged parts back together with the boundary
	loggedBody := boundary + "\r\n" + strings.Join(loggedParts, "\r\n"+boundary+"\r\n") + "\r\n" + boundary + "--"

	log.Info("Request body preview", zap.String("body", loggedBody))
}
