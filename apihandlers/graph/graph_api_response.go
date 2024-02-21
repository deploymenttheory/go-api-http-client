// graph_api_response.go
package graph

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
func (g *GraphAPIHandler) HandleAPISuccessResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	// Special handling for DELETE requests
	if resp.Request.Method == "DELETE" {
		return g.handleDeleteRequest(resp)
	}

	// Read the response body
	bodyBytes, err := g.readResponseBody(resp)
	if err != nil {
		return err
	}

	// Log the raw response details for debugging
	g.logResponseDetails(resp, bodyBytes)

	// Unmarshal the response based on content type
	contentType := resp.Header.Get("Content-Type")

	// Check for binary data handling
	contentDisposition := resp.Header.Get("Content-Disposition")
	if err := g.handleBinaryData(contentType, contentDisposition, bodyBytes, out); err != nil {
		return err
	}

	return g.unmarshalResponse(contentType, bodyBytes, out)
}

func (g *GraphAPIHandler) HandleAPIErrorResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	// Read the response body
	bodyBytes, err := g.readResponseBody(resp)
	if err != nil {
		return err
	}

	// Convert bodyBytes to a string to represent the raw response body
	rawResponse := string(bodyBytes)

	// Log the raw response details for debugging
	g.logResponseDetails(resp, bodyBytes)

	// Get the content type from the response headers
	contentType := resp.Header.Get("Content-Type")

	// Handle known error content types (e.g., JSON, HTML)
	if strings.Contains(contentType, "application/json") {
		return g.handleErrorJSONResponse(bodyBytes, resp.StatusCode, rawResponse)
	} else if strings.Contains(contentType, "text/html") {
		return g.handleErrorHTMLResponse(bodyBytes, resp.StatusCode)
	}

	// Generic error handling for unknown content types
	g.Logger.Error("Received non-success status code without detailed error response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("raw_response", rawResponse),
	)
	return fmt.Errorf("received non-success status code: %d, raw response: %s", resp.StatusCode, rawResponse)
}

// handleDeleteRequest handles the special case for DELETE requests, where a successful response might not contain a body.
func (g *GraphAPIHandler) handleDeleteRequest(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return g.Logger.Error("DELETE request failed", zap.Int("Status Code", resp.StatusCode))
}

// readResponseBody reads and returns the body of an HTTP response. It logs an error if reading fails.
func (g *GraphAPIHandler) readResponseBody(resp *http.Response) ([]byte, error) {
	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		g.Logger.Error("Failed reading response body", zap.Error(err))
		return nil, err
	}
	return bodyBytes, nil
}

// logResponseDetails logs the raw HTTP response body and headers for debugging purposes.
func (g *GraphAPIHandler) logResponseDetails(resp *http.Response, bodyBytes []byte) {
	// Log the response body as a string
	g.Logger.Debug("Raw HTTP Response", zap.String("Body", string(bodyBytes)))
	// Log the response headers
	g.Logger.Debug("HTTP Response Headers", zap.Any("Headers", resp.Header))
}

// handleBinaryData checks if the response should be treated as binary data based on the Content-Type or Content-Disposition headers. It assigns the response body to 'out' if 'out' is of type *[]byte.
func (g *GraphAPIHandler) handleBinaryData(contentType, contentDisposition string, bodyBytes []byte, out interface{}) error {
	// Check if response is binary data either by Content-Type or Content-Disposition
	if strings.Contains(contentType, "application/octet-stream") || strings.HasPrefix(contentDisposition, "attachment") {
		// Assert that 'out' is of the correct type to receive binary data
		if outPointer, ok := out.(*[]byte); ok {
			*outPointer = bodyBytes               // Assign the response body to 'out'
			g.Logger.Debug("Handled binary data", // Log handling of binary data
				zap.String("Content-Type", contentType),
				zap.String("Content-Disposition", contentDisposition),
			)
			return nil
		} else {
			errMsg := "output parameter is not a *[]byte for binary data"
			g.Logger.Error("Binary data handling error", // Log error for incorrect 'out' type
				zap.String("error", errMsg),
				zap.String("Content-Type", contentType),
				zap.String("Content-Disposition", contentDisposition),
			)
			return fmt.Errorf(errMsg)
		}
	}
	return nil // If not binary data, no action needed
}

// handleErrorHTMLResponse handles error responses with HTML content by extracting and logging the error message.
func (g *GraphAPIHandler) handleErrorHTMLResponse(bodyBytes []byte, statusCode int) error {
	// Extract the error message from the HTML content
	htmlErrorMessage := ExtractErrorMessageFromHTML(string(bodyBytes))
	// Log the error message along with the status code
	g.Logger.Error("Received HTML error content",
		zap.String("error_message", htmlErrorMessage),
		zap.Int("status_code", statusCode),
	)
	// Return an error with the extracted message
	return fmt.Errorf("received HTML error content: %s", htmlErrorMessage)
}

// handleErrorJSONResponse handles error responses with JSON content by parsing the error message and logging it.
func (g *GraphAPIHandler) handleErrorJSONResponse(bodyBytes []byte, statusCode int, rawResponse string) error {
	// Parse the JSON error response to extract the error description
	description, err := ParseJSONErrorResponse(bodyBytes)
	if err != nil {
		// Log the parsing error
		g.Logger.Error("Failed to parse JSON error response",
			zap.Error(err),
			zap.Int("status_code", statusCode),
			zap.String("raw_response", rawResponse), // Include raw response in the log
		)
		return fmt.Errorf("failed to parse JSON error response: %v, raw response: %s", err, rawResponse)
	}
	// Log the error description along with the status code and raw response
	g.Logger.Error("Received non-success status code with JSON response",
		zap.Int("status_code", statusCode),
		zap.String("error_description", description),
		zap.String("raw_response", rawResponse), // Include raw response in the log
	)
	return fmt.Errorf("received non-success status code with JSON response: %s, raw response: %s", description, rawResponse)
}

// unmarshalResponse unmarshals the response body into the provided output structure based on the content type (JSON or XML).
func (g *GraphAPIHandler) unmarshalResponse(contentType string, bodyBytes []byte, out interface{}) error {
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
		g.Logger.Error("Unmarshal error", zap.String("unexpected content type", contentType))
		return fmt.Errorf("unexpected content type: %s", contentType)
	}
}
