// response/success.go
/* Responsible for handling successful API responses. It reads the response body, logs the raw response details,
and unmarshals the response based on the content type (JSON or XML). */
package response

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// HandleAPISuccessResponse handles the HTTP success response from an API and unmarshals the response body into the provided output struct.
func HandleAPISuccessResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	// Special handling for DELETE requests
	if resp.Request.Method == "DELETE" {
		return handleDeleteRequest(resp, log)
	}

	// Read the response body
	bodyBytes, err := readResponseBody(resp, log)
	if err != nil {
		return err
	}

	// Log the raw response details for debugging
	logResponseDetails(resp, bodyBytes, log)

	// Unmarshal the response based on content type
	contentType := resp.Header.Get("Content-Type")

	// Check for binary data handling
	contentDisposition := resp.Header.Get("Content-Disposition")
	if err := handleBinaryData(contentType, contentDisposition, bodyBytes, log, out); err != nil {
		return err
	}

	return unmarshalResponse(contentType, bodyBytes, log, out)
}

// handleDeleteRequest handles the special case for DELETE requests, where a successful response might not contain a body.
func handleDeleteRequest(resp *http.Response, log logger.Logger) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Info("Successfully processed DELETE request", zap.String("URL", resp.Request.URL.String()), zap.Int("Status Code", resp.StatusCode))
		return nil
	}
	return log.Error("DELETE request failed", zap.String("URL", resp.Request.URL.String()), zap.Int("Status Code", resp.StatusCode))
}

// readResponseBody reads and returns the body of an HTTP response. It logs an error if reading fails.
func readResponseBody(resp *http.Response, log logger.Logger) ([]byte, error) {
	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed reading response body", zap.Error(err))
		return nil, err
	}
	return bodyBytes, nil
}

// logResponseDetails logs the raw HTTP response body and headers for debugging purposes.
func logResponseDetails(resp *http.Response, bodyBytes []byte, log logger.Logger) {
	// Log the response body as a string
	log.Debug("Raw HTTP Response", zap.String("Body", string(bodyBytes)))
	// Log the response headers
	log.Debug("HTTP Response Headers", zap.Any("Headers", resp.Header))
}

// handleBinaryData checks if the response should be treated as binary data based on the Content-Type or Content-Disposition headers. It assigns the response body to 'out' if 'out' is of type *[]byte.
func handleBinaryData(contentType, contentDisposition string, bodyBytes []byte, log logger.Logger, out interface{}) error {
	// Check if response is binary data either by Content-Type or Content-Disposition
	if strings.Contains(contentType, "application/octet-stream") || strings.HasPrefix(contentDisposition, "attachment") {
		// Assert that 'out' is of the correct type to receive binary data
		if outPointer, ok := out.(*[]byte); ok {
			*outPointer = bodyBytes          // Assign the response body to 'out'
			log.Debug("Handled binary data", // Log handling of binary data
				zap.String("Content-Type", contentType),
				zap.String("Content-Disposition", contentDisposition),
			)
			return nil
		} else {
			errMsg := "output parameter is not a *[]byte for binary data"
			log.Error("Binary data handling error", // Log error for incorrect 'out' type
				zap.String("error", errMsg),
				zap.String("Content-Type", contentType),
				zap.String("Content-Disposition", contentDisposition),
			)
			return fmt.Errorf(errMsg)
		}
	}
	return nil // If not binary data, no action needed
}

// unmarshalResponse unmarshals the response body into the provided output structure based on the MIME
// type extracted from the Content-Type header.
func unmarshalResponse(contentTypeHeader string, bodyBytes []byte, log logger.Logger, out interface{}) error {
	// Extract MIME type from Content-Type header
	mimeType, _ := ParseContentTypeHeader(contentTypeHeader)

	// Determine the MIME type and unmarshal accordingly
	switch {
	case strings.Contains(mimeType, "application/json"):
		// Unmarshal JSON content
		if err := json.Unmarshal(bodyBytes, out); err != nil {
			log.Error("JSON Unmarshal error", zap.Error(err))
			return err
		}
		log.Info("Successfully unmarshalled JSON response", zap.String("content type", mimeType))

	case strings.Contains(mimeType, "application/xml") || strings.Contains(mimeType, "text/xml"):
		// Unmarshal XML content
		if err := xml.Unmarshal(bodyBytes, out); err != nil {
			log.Error("XML Unmarshal error", zap.Error(err))
			return err
		}
		log.Info("Successfully unmarshalled XML response", zap.String("content type", mimeType))

	default:
		// Log and return an error for unexpected MIME types
		errMsg := fmt.Sprintf("unexpected MIME type: %s", mimeType)
		log.Error("Unmarshal error", zap.String("content type", mimeType), zap.Error(fmt.Errorf(errMsg)))
		return fmt.Errorf(errMsg)
	}
	return nil
}
