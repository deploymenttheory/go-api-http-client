// httpclient_error_response.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/logger"
)

// APIError represents a more flexible structure for API error responses.
type APIError struct {
	StatusCode int                    // HTTP status code
	Type       string                 // A brief identifier for the type of error
	Message    string                 // Human-readable message
	Detail     string                 // Detailed error message
	Errors     map[string]interface{} // A map to hold various error fields
	Raw        string                 // Raw response body for unstructured errors
}

// StructuredError represents a structured error response from the API.
type StructuredError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Error returns a string representation of the APIError.
func (e *APIError) Error() string {
	return fmt.Sprintf("API Error (Type: %s, Code: %d): %s", e.Type, e.StatusCode, e.Message)
}

// handleAPIErrorResponse attempts to parse the error response from the API and logs using zap logger.
func handleAPIErrorResponse(resp *http.Response, log logger.Logger) *APIError {
	// Initialize apiError with the HTTP status code and other fields
	apiError := &APIError{
		StatusCode: resp.StatusCode,
		Type:       "APIError",          // Default error type
		Message:    "An error occurred", // Default error message
	}

	// Read the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// Log and return an error if reading the body fails
		log.LogError("READ", resp.Request.URL.String(), resp.StatusCode, err, "Failed to read API error response body")
		return apiError
	}

	// Attempt to parse the response into a StructuredError
	if err := json.Unmarshal(bodyBytes, &apiError); err == nil && apiError.Message != "" {
		// Log the structured error with consistency
		log.LogError("API", resp.Request.URL.String(), resp.StatusCode, fmt.Errorf(apiError.Message), "")
		return apiError
	}

	// If structured parsing fails, attempt to parse into a generic error map
	var genericErr map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &genericErr); err == nil {
		// Extract fields from the generic error map and update apiError accordingly
		apiError.updateFromGenericError(genericErr)

		// Log the error with extracted details consistently
		log.LogError("API", resp.Request.URL.String(), resp.StatusCode, fmt.Errorf(apiError.Message), "")
		return apiError
	}

	// If all parsing attempts fail, log the raw response
	log.LogError("API", resp.Request.URL.String(), resp.StatusCode, fmt.Errorf("failed to parse API error response"), string(bodyBytes))
	return apiError
}

// updateFromGenericError updates the APIError fields based on a generic error map
func (e *APIError) updateFromGenericError(genericErr map[string]interface{}) {
	if msg, ok := genericErr["message"].(string); ok {
		e.Message = msg
	}
	if detail, ok := genericErr["detail"].(string); ok {
		e.Detail = detail
	}
	// Add more fields if necessary
}
