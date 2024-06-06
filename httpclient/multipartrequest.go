package httpclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
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
// This function handles constructing the multipart request body, setting the necessary headers, and executing the request.
// It supports custom content types and headers for each part of the multipart request, and handles authentication and
// logging throughout the process.

// Parameters:
// - method: A string representing the HTTP method to be used for the request. This method should be either POST or PUT
//   as these are the only methods that support multipart/form-data requests.
// - endpoint: The target API endpoint for the request. This should be a relative path that will be appended to the base URL
//   configured for the HTTP client.
// - files: A map where the key is the field name and the value is a slice of file paths to be included in the request.
// - formDataFields: A map of additional form fields to be included in the multipart request, where the key is the field name
//   and the value is the field value.
// - fileContentTypes: A map specifying the content type for each file part. The key is the field name and the value is the
//   content type (e.g., "image/jpeg").
// - formDataPartHeaders: A map specifying custom headers for each part of the multipart form data. The key is the field name
//   and the value is an http.Header containing the headers for that part.
// - out: A pointer to an output variable where the response will be deserialized. This should be a pointer to a struct that
//   matches the expected response schema.

// Returns:
// - *http.Response: The HTTP response received from the server. In case of successful execution, this response contains
//   the status code, headers, and body of the response. In case of errors, this response may contain the last received
//   HTTP response that led to the failure.
// - error: An error object indicating failure during request execution. This could be due to network issues, server errors,
//   or a failure in request serialization/deserialization.

// Usage:
// This function is suitable for executing multipart/form-data HTTP requests, particularly for file uploads along with
// additional form fields. It ensures proper authentication, sets necessary headers, and logs the process for debugging
// and monitoring purposes.

// Example:
// var result MyResponseType
// resp, err := client.DoMultiPartRequest("POST", "/api/upload", files, formDataFields, fileContentTypes, formDataPartHeaders, &result)
//
//	if err != nil {
//	    // Handle error
//	}
//
// // Use `result` or `resp` as needed
func (c *Client) DoMultiPartRequest(method, endpoint string, files map[string][]string, formDataFields map[string]string, fileContentTypes map[string]string, formDataPartHeaders map[string]http.Header, out interface{}) (*http.Response, error) {
	log := c.Logger

	if method != http.MethodPost && method != http.MethodPut {
		log.Error("HTTP method not supported for multipart request", zap.String("method", method))
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

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

	log.Info("Executing multipart request", zap.String("method", method), zap.String("endpoint", endpoint))

	// body, contentType, err := createMultipartRequestBody(files, formDataFields, fileContentTypes, formDataPartHeaders, log)
	// if err != nil {
	// 	return nil, err
	// }

	// Call the helper function to create a streaming multipart request body
	body, contentType, err := createStreamingMultipartRequestBody(files, formDataFields, fileContentTypes, formDataPartHeaders, log)
	if err != nil {
		return nil, err
	}

	//logMultiPartRequestBody(body, log)

	url := c.APIHandler.ConstructAPIResourceEndpoint(endpoint, log)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.clientConfig.ClientOptions.Timeout.CustomTimeout.Duration())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		log.Error("Failed to create HTTP request", zap.Error(err))
		return nil, err
	}

	cookiejar.ApplyCustomCookies(req, c.clientConfig.ClientOptions.Cookies.CustomCookies, c.Logger)

	req.Header.Set("Content-Type", contentType)

	headerHandler := headers.NewHeaderHandler(req, c.Logger, c.APIHandler, c.AuthTokenHandler)
	headerHandler.SetRequestHeaders(endpoint)
	headerHandler.LogHeaders(c.clientConfig.ClientOptions.Logging.HideSensitiveData)

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

// createStreamingMultipartRequestBody creates a streaming multipart request body with the provided files and form fields.
// This function constructs the body of a multipart/form-data request using an io.Pipe, allowing the request to be sent in chunks.

func createStreamingMultipartRequestBody(files map[string][]string, formDataFields map[string]string, fileContentTypes map[string]string, formDataPartHeaders map[string]http.Header, log logger.Logger) (io.Reader, string, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		for fieldName, filePaths := range files {
			for _, filePath := range filePaths {
				if err := addFilePart(writer, fieldName, filePath, fileContentTypes, formDataPartHeaders, log); err != nil {
					pw.CloseWithError(err)
					return
				}
			}
		}

		for key, val := range formDataFields {
			if err := addFormField(writer, key, val, log); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	return pr, writer.FormDataContentType(), nil
}

// createMultipartRequestBody creates a multipart request body with the provided files and form fields, supporting custom content types and headers.
// This function constructs the body of a multipart/form-data request by adding each file and form field to the multipart writer,
// setting custom content types and headers for each part as specified.

// Parameters:
// - files: A map where the key is the field name and the value is a slice of file paths to be included in the request. Each file path
//   corresponds to a file that will be included in the multipart request.
// - formDataFields: A map of additional form fields to be included in the multipart request, where the key is the field name
//   and the value is the field value. These are regular form fields that accompany the file uploads.
// - fileContentTypes: A map specifying the content type for each file part. The key is the field name and the value is the
//   content type (e.g., "image/jpeg"). If a content type is not specified for a field, "application/octet-stream" will be used as default.
// - formDataPartHeaders: A map specifying custom headers for each part of the multipart form data. The key is the field name
//   and the value is an http.Header containing the headers for that part. These headers are added to the multipart parts individually.
// - log: An instance of a logger implementing the logger.Logger interface, used to log informational messages, warnings,
//   and errors encountered during the construction of the multipart request body.

// Returns:
// - *bytes.Buffer: The constructed multipart request body. This buffer contains the full multipart form data payload ready to be sent.
// - string: The content type of the multipart request body. This includes the boundary string used by the multipart writer.
// - error: An error object indicating failure during the construction of the multipart request body. This could be due to issues
//   such as file reading errors or multipart writer errors.

func createMultipartRequestBody(files map[string][]string, formDataFields map[string]string, fileContentTypes map[string]string, formDataPartHeaders map[string]http.Header, log logger.Logger) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for fieldName, filePaths := range files {
		for _, filePath := range filePaths {
			if err := addFilePart(writer, fieldName, filePath, fileContentTypes, formDataPartHeaders, log); err != nil {
				return nil, "", err
			}
		}
	}

	for key, val := range formDataFields {
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
// This function opens the specified file, sets the appropriate content type and headers, and adds it to the multipart writer.

// Parameters:
// - writer: The multipart writer used to construct the multipart request body.
// - fieldName: The field name for the file part.
// - filePath: The path to the file to be included in the request.
// - fileContentTypes: A map specifying the content type for each file part. The key is the field name and the value is the
//   content type (e.g., "image/jpeg").
// - formDataPartHeaders: A map specifying custom headers for each part of the multipart form data. The key is the field name
//   and the value is an http.Header containing the headers for that part.
// - log: An instance of a logger implementing the logger.Logger interface, used to log informational messages, warnings,
//   and errors encountered during the addition of the file part.

// Returns:
//   - error: An error object indicating failure during the addition of the file part. This could be due to issues such as
//     file reading errors or multipart writer errors.
func addFilePart(writer *multipart.Writer, fieldName, filePath string, fileContentTypes map[string]string, formDataPartHeaders map[string]http.Header, log logger.Logger) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("Failed to open file", zap.String("filePath", filePath), zap.Error(err))
		return err
	}
	defer file.Close()

	// Default fileContentType
	contentType := "application/octet-stream"
	if ct, ok := fileContentTypes[fieldName]; ok {
		contentType = ct
	}

	header := setFormDataPartHeader(fieldName, filepath.Base(filePath), contentType, formDataPartHeaders[fieldName])

	part, err := writer.CreatePart(header)
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
// This function adds a regular form field (non-file) to the multipart request body.

// Parameters:
// - writer: The multipart writer used to construct the multipart request body.
// - key: The field name for the form field.
// - val: The value of the form field.
// - log: An instance of a logger implementing the logger.Logger interface, used to log informational messages, warnings,
//   and errors encountered during the addition of the form field.

// Returns:
//   - error: An error object indicating failure during the addition of the form field. This could be due to issues such as
//     multipart writer errors.
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

// setFormDataPartHeader creates a textproto.MIMEHeader for a form data field with the provided field name, file name, content type, and custom headers.
// This function constructs the MIME headers for a multipart form data part, including the content disposition, content type,
// and any custom headers specified.

// Parameters:
// - fieldname: The name of the form field.
// - filename: The name of the file being uploaded (if applicable).
// - contentType: The content type of the form data part (e.g., "image/jpeg").
// - customHeaders: A map of custom headers to be added to the form data part. The key is the header name and the value is the
//   header value.

// Returns:
// - textproto.MIMEHeader: The constructed MIME header for the form data part.
func setFormDataPartHeader(fieldname, filename, contentType string, customHeaders http.Header) textproto.MIMEHeader {
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldname, filename))
	header.Set("Content-Type", contentType)
	header.Set("Content-Transfer-Encoding", "base64")
	for key, values := range customHeaders {
		for _, value := range values {
			header.Add(key, value)
		}
	}
	return header
}

// chunkFileUpload reads the file upload into chunks and writes it to the writer.
// This function reads the file in chunks and writes it to the provided writer, allowing for progress logging during the upload.
// chunk size is set to 5124 KB (5 MB) by default.

// Parameters:
// - file: The file to be uploaded.
// - writer: The writer to which the file content will be written.
// - log: An instance of a logger implementing the logger.Logger interface, used to log informational messages, warnings,
//   and errors encountered during the file upload.
// - updateProgress: A function to update the upload progress, typically used for logging purposes.

// Returns:
//   - error: An error object indicating failure during the file upload. This could be due to issues such as file reading errors
//     or writer errors.
func chunkFileUpload(file *os.File, writer io.Writer, log logger.Logger, updateProgress func(int64)) error {
	const chunkSize = 10240 * 1024 // 5124 KB in bytes (5 MB)
	buffer := make([]byte, chunkSize)
	totalWritten := int64(0)
	chunkWritten := int64(0)
	fileName := filepath.Base(file.Name())

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

		if chunkWritten >= chunkSize {
			log.Debug("File Upload Chunk Sent",
				zap.String("file_name", fileName),
				zap.Int64("kb_sent", chunkWritten/1024),
				zap.Int64("total_kb_sent", totalWritten/1024))
			chunkWritten = 0
		}
	}

	// Log any remaining bytes that were written but didn't reach the log threshold
	if chunkWritten > 0 {
		log.Debug("Final Upload Chunk Sent",
			zap.String("file_name", fileName),
			zap.Int64("kb_sent", chunkWritten/1024),
			zap.Int64("total_kb_sent", totalWritten/1024))
	}

	return nil
}

// logUploadProgress logs the upload progress based on the percentage of the total file size.
// This function returns a closure that logs the upload progress each time it is called, updating the percentage completed.

// Parameters:
// - fileSize: The total size of the file being uploaded.
// - log: An instance of a logger implementing the logger.Logger interface, used to log informational messages, warnings,
//   and errors encountered during the upload.

// Returns:
// - func(int64): A function that takes the number of bytes written as an argument and logs the upload progress.
func logUploadProgress(fileSize int64, log logger.Logger) func(int64) {
	var uploaded int64 = 0
	const logInterval = 5 // Log every 5% increment
	lastLoggedPercentage := int64(0)

	return func(bytesWritten int64) {
		uploaded += bytesWritten
		percentage := (uploaded * 100) / fileSize

		if percentage >= lastLoggedPercentage+logInterval {
			log.Info("Upload progress",
				zap.Int64("uploaded_kbs", uploaded/1024),
				zap.Int64("total_filesize_in_kb", fileSize/1024),
				zap.String("percentage", fmt.Sprintf("%d%%", percentage)))
			lastLoggedPercentage = percentage
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
