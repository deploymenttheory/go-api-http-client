// response/success.go
/* Responsible for handling successful API responses. It reads the response body, logs the raw response details,
and unmarshals the response based on the content type (JSON or XML). */
package response

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// contentHandler defines the signature for unmarshaling content from an io.Reader.
type contentHandler func(io.Reader, interface{}, logger.Logger, string) error

// responseUnmarshallers maps MIME types to the corresponding contentHandler functions.
var responseUnmarshallers = map[string]contentHandler{
	"application/json": unmarshalJSON,
	"application/xml":  unmarshalXML,
	"text/xml":         unmarshalXML,
}

// HandleAPISuccessResponse reads the response body, logs the raw response details, and unmarshals the response based on the content type.
func HandleAPISuccessResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	if resp.Request.Method == "DELETE" {
		return handleDeleteRequest(resp, log)
	}

	// No need to read the entire body into memory, pass resp.Body directly.
	logResponseDetails(resp, nil, log) // Updated to handle nil bodyBytes.

	mimeType, _ := ParseContentTypeHeader(resp.Header.Get("Content-Type"))
	contentDisposition := resp.Header.Get("Content-Disposition")

	if handler, ok := responseUnmarshallers[mimeType]; ok {
		// Pass resp.Body directly to the handler for streaming.
		return handler(resp.Body, out, log, mimeType)
	} else if isBinaryData(mimeType, contentDisposition) {
		// For binary data, we still need to handle the body directly.
		return handleBinaryData(resp.Body, log, out, mimeType, contentDisposition)
	} else {
		errMsg := fmt.Sprintf("unexpected MIME type: %s", mimeType)
		log.Error("Unmarshal error", zap.String("content type", mimeType), zap.Error(errors.New(errMsg)))
		return errors.New(errMsg)
	}
}

// handleDeleteRequest handles the special case for DELETE requests, where a successful response might not contain a body.
func handleDeleteRequest(resp *http.Response, log logger.Logger) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if log != nil {
			log.Info("Successfully processed DELETE request", zap.String("URL", resp.Request.URL.String()), zap.Int("Status Code", resp.StatusCode))
		}
		return nil
	}
	if log != nil {
		return log.Error("DELETE request failed", zap.String("URL", resp.Request.URL.String()), zap.Int("Status Code", resp.StatusCode))
	}
	return fmt.Errorf("DELETE request failed, status code: %d", resp.StatusCode)
}

// Adjusted logResponseDetails to handle a potential nil bodyBytes.
func logResponseDetails(resp *http.Response, bodyBytes []byte, log logger.Logger) {
	// Conditional logging if bodyBytes is not nil.
	if bodyBytes != nil {
		log.Debug("Raw HTTP Response", zap.String("Body", string(bodyBytes)))
	}
	// Logging headers remains unchanged.
	log.Debug("HTTP Response Headers", zap.Any("Headers", resp.Header))
}

// unmarshalJSON unmarshals JSON content from an io.Reader into the provided output structure.
func unmarshalJSON(reader io.Reader, out interface{}, log logger.Logger, mimeType string) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(out); err != nil {
		log.Error("JSON Unmarshal error", zap.Error(err))
		return err
	}
	log.Info("Successfully unmarshalled JSON response", zap.String("content type", mimeType))
	return nil
}

// unmarshalXML unmarshals XML content from an io.Reader into the provided output structure.
func unmarshalXML(reader io.Reader, out interface{}, log logger.Logger, mimeType string) error {
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(out); err != nil {
		log.Error("XML Unmarshal error", zap.Error(err))
		return err
	}
	log.Info("Successfully unmarshalled XML response", zap.String("content type", mimeType))
	return nil
}

// isBinaryData checks if the MIME type or Content-Disposition indicates binary data.
func isBinaryData(contentType, contentDisposition string) bool {
	return strings.Contains(contentType, "application/octet-stream") || strings.HasPrefix(contentDisposition, "attachment")
}

// handleBinaryData reads binary data from an io.Reader and stores it in *[]byte or streams it to an io.Writer.
func handleBinaryData(reader io.Reader, log logger.Logger, out interface{}, mimeType, contentDisposition string) error {
	// Check if the output interface is either *[]byte or io.Writer
	switch out := out.(type) {
	case *[]byte:
		// Read all data from reader and store it in *[]byte
		data, err := io.ReadAll(reader)
		if err != nil {
			log.Error("Failed to read binary data", zap.Error(err))
			return err
		}
		*out = data

	case io.Writer:
		// Stream data directly to the io.Writer
		_, err := io.Copy(out, reader)
		if err != nil {
			log.Error("Failed to stream binary data to io.Writer", zap.Error(err))
			return err
		}

	default:
		errMsg := "output parameter is not suitable for binary data (*[]byte or io.Writer)"
		log.Error(errMsg, zap.String("Content-Type", mimeType))
		return errors.New(errMsg)
	}

	// Handle Content-Disposition if present
	if contentDisposition != "" {
		_, params := ParseContentDisposition(contentDisposition)
		if filename, ok := params["filename"]; ok {
			log.Debug("Extracted filename from Content-Disposition", zap.String("filename", filename))
			// Additional processing for the filename can be done here if needed
		}
	}

	return nil
}
