// jamfpro_api_handler.go
package jamfpro

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

// Functions
func (j *JamfAPIHandler) HandleAPISuccessResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	// Special handling for DELETE requests
	if resp.Request.Method == "DELETE" {
		return j.handleDeleteRequest(resp)
	}

	// Read the response body
	bodyBytes, err := j.readResponseBody(resp)
	if err != nil {
		return err
	}

	// Log the raw response details for debugging
	j.logResponseDetails(resp, bodyBytes)

	// Unmarshal the response based on content type
	contentType := resp.Header.Get("Content-Type")

	// Check for binary data handling
	contentDisposition := resp.Header.Get("Content-Disposition")
	if err := j.handleBinaryData(contentType, contentDisposition, bodyBytes, out); err != nil {
		return err
	}

	return j.unmarshalResponse(contentType, bodyBytes, out)
}

// handleDeleteRequest handles the special case for DELETE requests, where a successful response might not contain a body.
func (j *JamfAPIHandler) handleDeleteRequest(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return j.Logger.Error("DELETE request failed", zap.Int("Status Code", resp.StatusCode))
}

// readResponseBody reads and returns the body of an HTTP response. It logs an error if reading fails.
func (j *JamfAPIHandler) readResponseBody(resp *http.Response) ([]byte, error) {
	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		j.Logger.Error("Failed reading response body", zap.Error(err))
		return nil, err
	}
	return bodyBytes, nil
}

// logResponseDetails logs the raw HTTP response body and headers for debugging purposes.
func (j *JamfAPIHandler) logResponseDetails(resp *http.Response, bodyBytes []byte) {
	// Log the response body as a string
	j.Logger.Debug("Raw HTTP Response", zap.String("Body", string(bodyBytes)))
	// Log the response headers
	j.Logger.Debug("HTTP Response Headers", zap.Any("Headers", resp.Header))
}

// handleBinaryData checks if the response should be treated as binary data based on the Content-Type or Content-Disposition headers. It assigns the response body to 'out' if 'out' is of type *[]byte.
func (j *JamfAPIHandler) handleBinaryData(contentType, contentDisposition string, bodyBytes []byte, out interface{}) error {
	// Check if response is binary data either by Content-Type or Content-Disposition
	if strings.Contains(contentType, "application/octet-stream") || strings.HasPrefix(contentDisposition, "attachment") {
		// Assert that 'out' is of the correct type to receive binary data
		if outPointer, ok := out.(*[]byte); ok {
			*outPointer = bodyBytes               // Assign the response body to 'out'
			j.Logger.Debug("Handled binary data", // Log handling of binary data
				zap.String("Content-Type", contentType),
				zap.String("Content-Disposition", contentDisposition),
			)
			return nil
		} else {
			errMsg := "output parameter is not a *[]byte for binary data"
			j.Logger.Error("Binary data handling error", // Log error for incorrect 'out' type
				zap.String("error", errMsg),
				zap.String("Content-Type", contentType),
				zap.String("Content-Disposition", contentDisposition),
			)
			return fmt.Errorf(errMsg)
		}
	}
	return nil // If not binary data, no action needed
}

// unmarshalResponse unmarshals the response body into the provided output structure based on the content type (JSON or XML).
func (j *JamfAPIHandler) unmarshalResponse(contentType string, bodyBytes []byte, out interface{}) error {
	// Determine the content type and unmarshal accordingly
	switch {
	case strings.Contains(contentType, "application/json"):
		// Unmarshal JSON content
		return json.Unmarshal(bodyBytes, out)
	case strings.Contains(contentType, "application/xml"), strings.Contains(contentType, "text/xml;charset=UTF-8"):
		// Unmarshal XML content
		return xml.Unmarshal(bodyBytes, out)
	default:
		// Log and return an error for unexpected content types
		j.Logger.Error("Unmarshal error", zap.String("unexpected content type", contentType))
		return fmt.Errorf("unexpected content type: %s", contentType)
	}
}