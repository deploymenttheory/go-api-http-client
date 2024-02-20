// jamfpro_api_handler.go
package jamfpro

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// EndpointConfig is a struct that holds configuration details for a specific API endpoint.
// It includes what type of content it can accept and what content type it should send.
type EndpointConfig struct {
	Accept      string  `json:"accept"`       // Accept specifies the MIME type the endpoint can handle in responses.
	ContentType *string `json:"content_type"` // ContentType, if not nil, specifies the MIME type to set for requests sent to the endpoint. A pointer is used to distinguish between a missing field and an empty string.
}

// JamfAPIHandler implements the APIHandler interface for the Jamf Pro API.
type JamfAPIHandler struct {
	OverrideBaseDomain string        // OverrideBaseDomain is used to override the base domain for URL construction.
	InstanceName       string        // InstanceName is the name of the Jamf instance.
	Logger             logger.Logger // Logger is the structured logger used for logging.
}

// Functions

// MarshalRequest encodes the request body according to the endpoint for the API.
func (j *JamfAPIHandler) MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	// Determine the format based on the endpoint
	format := "json"
	if strings.Contains(endpoint, "/JSSResource") {
		format = "xml"
	} else if strings.Contains(endpoint, "/api") {
		format = "json"
	}

	switch format {
	case "xml":
		data, err = xml.Marshal(body)
		if err != nil {
			return nil, err
		}

		if method == "POST" || method == "PUT" {
			j.Logger.Debug("XML Request Body", zap.String("Body", string(data)))
		}

	case "json":
		data, err = json.Marshal(body)
		if err != nil {
			j.Logger.Error("Failed marshaling JSON request", zap.Error(err))
			return nil, err
		}

		if method == "POST" || method == "PUT" || method == "PATCH" {
			j.Logger.Debug("JSON Request Body", zap.String("Body", string(data)))
		}
	}

	return data, nil
}

// UnmarshalResponse decodes the response body from XML or JSON format depending on the Content-Type header.
func (j *JamfAPIHandler) UnmarshalResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	// Handle DELETE method
	if resp.Request.Method == "DELETE" {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		} else {
			return j.Logger.Error("DELETE request failed", zap.Int("Status Code", resp.StatusCode))
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		j.Logger.Error("Failed reading response body", zap.Error(err))
		return err
	}

	// Log the raw response body and headers
	j.Logger.Debug("Raw HTTP Response", zap.String("Body", string(bodyBytes)))
	j.Logger.Debug("Unmarshaling response", zap.String("status", resp.Status))

	// Log headers when in debug mode
	j.Logger.Debug("HTTP Response Headers", zap.Any("Headers", resp.Header))

	// Check the Content-Type and Content-Disposition headers
	contentType := resp.Header.Get("Content-Type")
	contentDisposition := resp.Header.Get("Content-Disposition")

	// Handle binary data if necessary
	if err := j.handleBinaryData(contentType, contentDisposition, bodyBytes, out); err != nil {
		return err
	}

	// Check for non-success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// If the content type is HTML, extract and log the error message
		if strings.Contains(contentType, "text/html") {
			htmlErrorMessage := ExtractErrorMessageFromHTML(string(bodyBytes))

			// Log the HTML error message using Zap
			j.Logger.Error("Received HTML error content",
				zap.String("error_message", htmlErrorMessage),
				zap.Int("status_code", resp.StatusCode),
			)
		} else {
			// Log a generic error message if the response is not HTML
			j.Logger.Error("Received non-success status code without detailed error response",
				zap.Int("status_code", resp.StatusCode),
			)
		}
	}

	// Check for non-success status codes before attempting to unmarshal
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse the error details from the response body for JSON content type
		if strings.Contains(contentType, "application/json") {
			description, err := ParseJSONErrorResponse(bodyBytes)
			if err != nil {
				// Log the error using the structured logger and return the error
				j.Logger.Error("Failed to parse JSON error response",
					zap.Error(err),
					zap.Int("status_code", resp.StatusCode),
				)
				return err
			}
			// Log the error with description using the structured logger and return the error
			j.Logger.Error("Received non-success status code with JSON response",
				zap.Int("status_code", resp.StatusCode),
				zap.String("error_description", description),
			)
			return fmt.Errorf("received non-success status code with JSON response: %s", description)
		}

		// If the response is not JSON or another error occurs, log a generic error message and return an error
		j.Logger.Error("Received non-success status code without JSON response",
			zap.Int("status_code", resp.StatusCode),
		)
		return fmt.Errorf("received non-success status code without JSON response: %d", resp.StatusCode)
	}

	// Determine whether the content type is JSON or XML and unmarshal accordingly
	switch {
	case strings.Contains(contentType, "application/json"):
		err = json.Unmarshal(bodyBytes, out)
	case strings.Contains(contentType, "application/xml"), strings.Contains(contentType, "text/xml;charset=UTF-8"):
		err = xml.Unmarshal(bodyBytes, out)
	default:
		// If the content type is neither JSON nor XML nor HTML
		return fmt.Errorf("unexpected content type: %s", contentType)
	}

	// Handle any errors that occurred during unmarshaling
	if err != nil {
		// If unmarshalling fails, check if the content might be HTML
		if strings.Contains(string(bodyBytes), "<html>") {
			htmlErrorMessage := ExtractErrorMessageFromHTML(string(bodyBytes))

			// Log the HTML error message
			j.Logger.Warn("Received HTML content instead of expected format",
				zap.String("error_message", htmlErrorMessage),
				zap.Int("status_code", resp.StatusCode),
			)

			// Use the HTML error message for logging the error
			j.Logger.Error("Unmarshal error with HTML content",
				zap.String("error_message", htmlErrorMessage),
				zap.Int("status_code", resp.StatusCode),
			)
		} else {
			// If the error is not due to HTML content, log the original error
			j.Logger.Error("Unmarshal error",
				zap.Error(err),
				zap.Int("status_code", resp.StatusCode),
			)
		}
	}

	return err
}

// MarshalMultipartFormData takes a map with form fields and file paths and returns the encoded body and content type.
func (j *JamfAPIHandler) MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the simple fields to the form data
	for field, value := range fields {
		if err := writer.WriteField(field, value); err != nil {
			return nil, "", err
		}
	}

	// Add the files to the form data
	for formField, filepath := range files {
		file, err := os.Open(filepath)
		if err != nil {
			return nil, "", err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(formField, filepath)
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(part, file); err != nil {
			return nil, "", err
		}
	}

	// Close the writer before returning
	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body.Bytes(), contentType, nil
}

// handleBinaryData checks if the response should be treated as binary data and assigns to out if so.
func (j *JamfAPIHandler) handleBinaryData(contentType, contentDisposition string, bodyBytes []byte, out interface{}) error {
	if strings.Contains(contentType, "application/octet-stream") || strings.HasPrefix(contentDisposition, "attachment") {
		if outPointer, ok := out.(*[]byte); ok {
			*outPointer = bodyBytes
			return nil
		} else {
			return fmt.Errorf("output parameter is not a *[]byte for binary data")
		}
	}
	return nil // If not binary data, no action needed
}
