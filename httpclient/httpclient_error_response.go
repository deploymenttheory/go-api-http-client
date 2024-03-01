// httpclient_error_response.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// Error returns a JSON representation of the APIError.
func (e *APIError) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf("Error encoding APIError to JSON: %s", err)
	}
	return string(data)
}

// handleAPIErrorResponse attempts to parse the error response from the API and logs using the zap logger.
func handleAPIErrorResponse(resp *http.Response, log logger.Logger) *APIError {
	apiError := &APIError{
		StatusCode: resp.StatusCode,
		Type:       "APIError",          // Default error type
		Message:    "An error occurred", // Default error message
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiError
	}

	// Check if the response is JSON
	if isJSONResponse(resp) {
		// Attempt to parse the response into a StructuredError
		if err := json.Unmarshal(bodyBytes, &apiError); err == nil && apiError.Message != "" {
			log.LogError(
				"json_structured_error_detected", // event
				resp.Request.Method,              // method
				resp.Request.URL.String(),        // url
				resp.StatusCode,                  // statusCode
				resp.Status,                      // status
				fmt.Errorf(apiError.Message),     // err
				apiError.Raw,                     // raw resp
			)
			return apiError
		}

		// If structured parsing fails, attempt to parse into a generic error map
		var genericErr map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &genericErr); err == nil {
			apiError.updateFromGenericError(genericErr)
			log.LogError(
				"json_generic_error_detected", // event
				resp.Request.Method,           // method
				resp.Request.URL.String(),     // url
				resp.StatusCode,               // statusCode
				resp.Status,                   // status
				fmt.Errorf(apiError.Message),  // err
				apiError.Raw,                  // raw resp
			)
			return apiError
		}
	} else if isHTMLResponse(resp) {
		// Handle HTML response
		apiError.Raw = string(bodyBytes)
		log.LogError(
			"api_html_error",             // event
			resp.Request.Method,          // method
			resp.Request.URL.String(),    // url
			resp.StatusCode,              // statusCode
			resp.Status,                  // status
			fmt.Errorf(apiError.Message), // err
			apiError.Raw,                 // raw resp
		)
		return apiError
	} else {
		// Handle other non-JSON responses
		apiError.Raw = string(bodyBytes)
		log.LogError(
			"api_non_json_error",      // event
			resp.Request.Method,       // method
			resp.Request.URL.String(), // url
			resp.StatusCode,           // statusCode
			resp.Status,               // status
			fmt.Errorf("Non-JSON error response received"), // err
			apiError.Raw, // raw resp
		)
		return apiError
	}

	return apiError
}

// isJSONResponse checks if the response Content-Type indicates JSON
func isJSONResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json")
}

// isHTMLResponse checks if the response Content-Type indicates HTML
func isHTMLResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "text/html")
}

// updateFromGenericError updates the APIError fields based on a generic error map extracted from an API response.
// This function is useful for cases where the error response does not match the predefined StructuredError format,
// and instead, a more generic error handling approach is needed. It extracts known fields such as 'message' and 'detail'
// from the generic error map and updates the corresponding fields in the APIError instance.
//
// Parameters:
// - genericErr: A map[string]interface{} representing the generic error structure extracted from an API response.
//
// The function checks for the presence of 'message' and 'detail' keys in the generic error map. If these keys are present,
// their values are used to update the 'Message' and 'Detail' fields of the APIError instance, respectively.
func (e *APIError) updateFromGenericError(genericErr map[string]interface{}) {
	if msg, ok := genericErr["message"].(string); ok {
		e.Message = msg
	}
	if detail, ok := genericErr["detail"].(string); ok {
		e.Detail = detail
	}
	// Optionally add more fields if necessary
}
