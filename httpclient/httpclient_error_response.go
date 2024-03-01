// httpclient_error_response.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
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
	apiError := &APIError{StatusCode: resp.StatusCode}

	// Attempt to parse the response into a StructuredError
	var structuredErr StructuredError
	if err := json.NewDecoder(resp.Body).Decode(&structuredErr); err == nil && structuredErr.Error.Message != "" {
		apiError.Type = structuredErr.Error.Code
		apiError.Message = structuredErr.Error.Message

		// Log the structured error details with zap logger
		log.Warn("API returned structured error",
			zap.String("error_code", structuredErr.Error.Code),
			zap.String("error_message", structuredErr.Error.Message),
			zap.Int("status_code", resp.StatusCode),
		)

		return apiError
	}

	// If the structured error parsing fails, attempt a more generic parsing
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// If reading the response body fails, store the error message and log the error
		apiError.Raw = "Failed to read API error response body"
		apiError.Message = err.Error()
		apiError.Type = "ReadError"

		log.Error("Failed to read API error response body",
			zap.Error(err),
		)

		return apiError
	}

	if err := json.Unmarshal(bodyBytes, &apiError.Errors); err != nil {
		// If generic parsing also fails, store the raw response body and log the error
		apiError.Raw = string(bodyBytes)
		apiError.Message = "Failed to parse API error response"
		apiError.Type = "UnexpectedError"

		log.Error("Failed to parse API error response",
			zap.String("raw_response", apiError.Raw),
		)

		return apiError
	}

	// Extract fields from the generic error map and log the error with extracted details
	if msg, ok := apiError.Errors["message"].(string); ok {
		apiError.Message = msg
	}
	if detail, ok := apiError.Errors["detail"].(string); ok {
		apiError.Detail = detail
	}

	log.Error("API error",
		zap.Int("status_code", apiError.StatusCode),
		zap.String("type", apiError.Type),
		zap.String("message", apiError.Message),
		zap.String("detail", apiError.Detail),
	)

	return apiError
}
