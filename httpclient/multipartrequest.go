package httpclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/deploymenttheory/go-api-http-client/response"
	"go.uber.org/zap"
)

// UploadState represents the state of an upload operation, including the last uploaded byte.
// This struct is used to track the progress of file uploads for resumable uploads and to resume uploads from the last uploaded byte.
type UploadState struct {
	LastUploadedByte int64
	sync.Mutex
}

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

	if method != http.MethodPost && method != http.MethodPut {
		c.Sugar.Error("HTTP method not supported for multipart request", zap.String("method", method))
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	url := (*c.Integration).GetFQDN() + endpoint

	var ctx context.Context
	var cancel context.CancelFunc

	if c.config.CustomTimeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), c.config.CustomTimeout)
		c.Sugar.Info("Using timeout context for multipart request", zap.Duration("custom_timeout_seconds", c.config.CustomTimeout))
	} else {
		ctx = context.Background()
		cancel = func() {}
		c.Sugar.Info("Using background context for multipart request. Caller will handle timeouts")
	}
	defer cancel()

	var body io.Reader
	var contentType string

	// Create multipart body in a function to ensure it runs again on retry
	createBody := func() error {
		var err error
		body, contentType, err = createStreamingMultipartRequestBody(files, formDataFields, fileContentTypes, formDataPartHeaders, c.Sugar)
		if err != nil {
			c.Sugar.Error("Failed to create streaming multipart request body", zap.Error(err))
		} else {
			c.Sugar.Info("Successfully created streaming multipart request body", zap.String("content_type", contentType))
		}
		return err
	}

	if err := createBody(); err != nil {
		c.Sugar.Error("Failed to create streaming multipart request body", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		c.Sugar.Error("Failed to create HTTP request", zap.Error(err))
		return nil, err
	}

	// sugar the request details
	c.Sugar.Info("Created HTTP Multipart request", zap.String("method", method), zap.String("url", url), zap.String("content_type", contentType))

	(*c.Integration).PrepRequestParamsAndAuth(req)

	req.Header.Set("Content-Type", contentType)

	startTime := time.Now()
	resp, requestErr := c.http.Do(req)
	duration := time.Since(startTime)

	if requestErr != nil {
		c.Sugar.Error("Failed to send request", zap.String("method", method), zap.String("endpoint", endpoint), zap.Error(requestErr))
		return nil, requestErr
	}

	c.Sugar.Debug("Request sent successfully", zap.String("method", method), zap.String("endpoint", endpoint), zap.Int("status_code", resp.StatusCode), zap.Duration("duration", duration))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return resp, response.HandleAPISuccessResponse(resp, out, c.Sugar)
	}

	return resp, response.HandleAPIErrorResponse(resp, c.Sugar)
}

// createStreamingMultipartRequestBody creates a streaming multipart request body with the provided files and form fields.
// This function constructs the body of a multipart/form-data request using an io.Pipe, allowing the request to be sent in chunks.
// It supports custom content types and headers for each part of the multipart request, and logs the process for debugging
// and monitoring purposes.

// Parameters:
// - files: A map where the key is the field name and the value is a slice of file paths to be included in the request.
//   Each file path corresponds to a file that will be included in the multipart request.
// - formDataFields: A map of additional form fields to be included in the multipart request, where the key is the field name
//   and the value is the field value. These are regular form fields that accompany the file uploads.
// - fileContentTypes: A map specifying the content type for each file part. The key is the field name and the value is the
//   content type (e.g., "image/jpeg").
// - formDataPartHeaders: A map specifying custom headers for each part of the multipart form data. The key is the field name
//   and the value is an http.Header containing the headers for that part.
// - sugar: An instance of a logger implementing the logger.Logger interface, used to sugar informational messages, warnings,
//   and errors encountered during the construction of the multipart request body.

// Returns:
//   - io.Reader: The constructed multipart request body reader. This reader streams the multipart form data payload ready to be sent.
//   - string: The content type of the multipart request body. This includes the boundary string used by the multipart writer.
//   - error: An error object indicating failure during the construction of the multipart request body. This could be due to issues
//     such as file reading errors or multipart writer errors.
func createStreamingMultipartRequestBody(files map[string][]string, formDataFields map[string]string, fileContentTypes map[string]string, formDataPartHeaders map[string]http.Header, sugar *zap.SugaredLogger) (io.Reader, string, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer func() {
			if err := writer.Close(); err != nil {
				sugar.Error("Failed to close multipart writer", zap.Error(err))
			}
			if err := pw.Close(); err != nil {
				sugar.Error("Failed to close pipe writer", zap.Error(err))
			}
		}()

		for fieldName, filePaths := range files {
			for _, filePath := range filePaths {
				sugar.Debug("Adding file part", zap.String("field_name", fieldName), zap.String("file_path", filePath))
				if err := addFilePart(writer, fieldName, filePath, fileContentTypes, formDataPartHeaders, sugar); err != nil {
					sugar.Error("Failed to add file part", zap.Error(err))
					pw.CloseWithError(err)
					return
				}
			}
		}

		for key, val := range formDataFields {
			sugar.Debug("Adding form field", zap.String("field_name", key), zap.String("field_value", val))
			if err := addFormField(writer, key, val, sugar); err != nil {
				sugar.Error("Failed to add form field", zap.Error(err))
				pw.CloseWithError(err)
				return
			}
		}
	}()

	return pr, writer.FormDataContentType(), nil
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
// - sugar: An instance of a logger implementing the logger.Logger interface, used to sugar informational messages, warnings,
//   and errors encountered during the addition of the file part.

// Returns:
//   - error: An error object indicating failure during the addition of the file part. This could be due to issues such as
//     file reading errors or multipart writer errors.
func addFilePart(writer *multipart.Writer, fieldName, filePath string, fileContentTypes map[string]string, formDataPartHeaders map[string]http.Header, sugar *zap.SugaredLogger) error {
	file, err := os.Open(filePath)
	if err != nil {
		sugar.Error("Failed to open file", zap.String("filePath", filePath), zap.Error(err))
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
		sugar.Error("Failed to create form file part", zap.String("fieldName", fieldName), zap.Error(err))
		return err
	}

	encoder := base64.NewEncoder(base64.StdEncoding, part)
	defer encoder.Close()

	fileSize, err := file.Stat()
	if err != nil {
		sugar.Error("Failed to get file info", zap.String("filePath", filePath), zap.Error(err))
		return err
	}

	progressLogger := logUploadProgress(file, fileSize.Size(), sugar)
	uploadState := &UploadState{}
	if err := chunkFileUpload(file, encoder, progressLogger, uploadState, sugar); err != nil {
		sugar.Error("Failed to copy file content", zap.String("filePath", filePath), zap.Error(err))
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
// - sugar: An instance of a logger implementing the logger.Logger interface, used to sugar informational messages, warnings,
//   and errors encountered during the addition of the form field.

// Returns:
//   - error: An error object indicating failure during the addition of the form field. This could be due to issues such as
//     multipart writer errors.
func addFormField(writer *multipart.Writer, key, val string, sugar *zap.SugaredLogger) error {
	fieldWriter, err := writer.CreateFormField(key)
	if err != nil {
		sugar.Error("Failed to create form field", zap.String("key", key), zap.Error(err))
		return err
	}
	if _, err := fieldWriter.Write([]byte(val)); err != nil {
		sugar.Error("Failed to write form field", zap.String("key", key), zap.Error(err))
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
// The chunk size is set to 8192 KB (8 MB) by default. This is a common chunk size used for file uploads to cloud storage services.

// Azure Blob Storage has a minimum chunk size of 4 MB and a maximum of 100 MB for block blobs.
// GCP Cloud Storage has a minimum chunk size of 256 KB and a maximum of 5 GB.
// AWS S3 has a minimum chunk size of 5 MB and a maximum of 5 GB.

// The function also calculates the total number of chunks and logs the chunk number during the upload process.

// Parameters:
// - file: The file to be uploaded.
// - writer: The writer to which the file content will be written.
// - sugar: An instance of a logger implementing the logger.Logger interface, used to sugar informational messages, warnings,
//   and errors encountered during the file upload.
// - updateProgress: A function to update the upload progress, typically used for logging purposes.
// - uploadState: A pointer to an UploadState struct used to track the progress of the file upload for resumable uploads.

// Returns:
//   - error: An error object indicating failure during the file upload. This could be due to issues such as file reading errors
//     or writer errors.
func chunkFileUpload(file *os.File, writer io.Writer, updateProgress func(int64), uploadState *UploadState, sugar *zap.SugaredLogger) error {
	const chunkSize = 8 * 1024 * 1024 // 8 MB
	buffer := make([]byte, chunkSize)
	totalWritten := int64(0)
	chunkWritten := int64(0)
	fileName := filepath.Base(file.Name())

	// Seek to the last uploaded byte
	file.Seek(uploadState.LastUploadedByte, io.SeekStart)

	// Calculate the total number of chunks
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	totalChunks := (fileInfo.Size() + chunkSize - 1) / chunkSize
	currentChunk := uploadState.LastUploadedByte / chunkSize

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
			// Save the state before returning the error
			uploadState.Lock()
			uploadState.LastUploadedByte += totalWritten
			uploadState.Unlock()
			return err
		}

		totalWritten += int64(written)
		chunkWritten += int64(written)
		updateProgress(int64(written))

		if chunkWritten >= chunkSize {
			currentChunk++
			sugar.Debug("File Upload Chunk Sent",
				zap.String("file_name", fileName),
				zap.Int64("chunk_number", currentChunk),
				zap.Int64("total_chunks", totalChunks),
				zap.Int64("kb_sent", chunkWritten/1024),
				zap.Int64("total_kb_sent", totalWritten/1024))
			chunkWritten = 0
		}
	}

	// sugar any remaining bytes that were written but didn't reach the sugar threshold
	if chunkWritten > 0 {
		currentChunk++
		sugar.Debug("Final Upload Chunk Sent",
			zap.String("file_name", fileName),
			zap.Int64("chunk_number", currentChunk),
			zap.Int64("total_chunks", totalChunks),
			zap.Int64("kb_sent", chunkWritten/1024),
			zap.Int64("total_kb_sent", totalWritten/1024))
	}

	return nil
}

// logUploadProgress logs the upload progress based on the percentage of the total file size.
// This function returns a closure that logs the upload progress each time it is called, updating the percentage completed.

// Parameters:
// - file: The file being uploaded. used for logging the file name.
// - fileSize: The total size of the file being uploaded.
// - sugar: An instance of a logger implementing the logger.Logger interface, used to sugar informational messages, warnings,
//   and errors encountered during the upload.

// Returns:
// - func(int64): A function that takes the number of bytes written as an argument and logs the upload progress.
// logUploadProgress logs the upload progress based on the percentage of the total file size.
func logUploadProgress(file *os.File, fileSize int64, sugar *zap.SugaredLogger) func(int64) {
	var uploaded int64 = 0
	const logInterval = 5 // sugar every 5% increment
	lastLoggedPercentage := int64(0)
	startTime := time.Now()
	fileName := filepath.Base(file.Name())

	return func(bytesWritten int64) {
		uploaded += bytesWritten
		percentage := (uploaded * 100) / fileSize

		if percentage >= lastLoggedPercentage+logInterval {
			elapsedTime := time.Since(startTime)

			sugar.Info("Upload progress",
				zap.String("file_name", fileName),
				zap.Float64("uploaded_MB's", float64(uploaded)/1048576), // sugar in MB (1024 * 1024)
				zap.Float64("total_filesize_in_MB", float64(fileSize)/1048576),
				zap.String("total_uploaded_percentage", fmt.Sprintf("%d%%", percentage)),
				zap.Duration("elapsed_time", elapsedTime))
			lastLoggedPercentage = percentage
		}
	}
}
