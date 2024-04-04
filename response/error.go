// response/error.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package response

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"golang.org/x/net/html"
)

// APIError represents an api error response.
type APIError struct {
	StatusCode int                    `json:"status_code" xml:"StatusCode"`            // HTTP status code
	Type       string                 `json:"type" xml:"Type"`                         // Type of error
	Message    string                 `json:"message" xml:"Message"`                   // Human-readable message
	Detail     string                 `json:"detail,omitempty" xml:"Detail,omitempty"` // Detailed error message
	Errors     map[string]interface{} `json:"errors,omitempty" xml:"Errors,omitempty"` // Additional error details
	Raw        string                 `json:"raw" xml:"Raw"`                           // Raw response body for debugging
}

// Error returns a string representation of the APIError, making it compatible with the error interface.
func (e *APIError) Error() string {
	// Attempt to marshal the APIError instance into a JSON string.
	data, err := json.Marshal(e)
	if err == nil {
		return string(data)
	}

	// Use the standard HTTP status text as the error message if 'Message' field is empty.
	if e.Message == "" {
		e.Message = http.StatusText(e.StatusCode)
	}

	// Fallback to a simpler error message format if JSON marshaling fails.
	return fmt.Sprintf("API Error: StatusCode=%d, Type=%s, Message=%s", e.StatusCode, e.Type, e.Message)
}

// HandleAPIErrorResponse handles the HTTP error response from an API and logs the error.
func HandleAPIErrorResponse(resp *http.Response, log logger.Logger) *APIError {
	apiError := &APIError{
		StatusCode: resp.StatusCode,
		Type:       "APIError",
		Message:    "An error occurred",
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		apiError.Raw = "Failed to read response body"
		logError(log, apiError, "error_reading_response_body", resp)
		return apiError
	}

	mimeType, _ := ParseContentTypeHeader(resp.Header.Get("Content-Type"))
	switch mimeType {
	case "application/json":
		parseJSONResponse(bodyBytes, apiError, log, resp)
		logError(log, apiError, "json_error_detected", resp)
	case "application/xml", "text/xml":
		parseXMLResponse(bodyBytes, apiError, log, resp)
		logError(log, apiError, "xml_error_detected", resp)
	case "text/html":
		parseHTMLResponse(bodyBytes, apiError, log, resp)
		logError(log, apiError, "html_error_detected", resp)
	case "text/plain":
		parseTextResponse(bodyBytes, apiError, log, resp)
		logError(log, apiError, "text_error_detected", resp)
	default:
		apiError.Raw = string(bodyBytes)
		apiError.Message = "Unknown content type error"
		logError(log, apiError, "unknown_content_type_error", resp)
	}

	return apiError
}

// ParseContentTypeHeader parses the Content-Type header and extracts the MIME type and parameters.
func ParseContentTypeHeader(header string) (string, map[string]string) {
	parts := strings.Split(header, ";")
	mimeType := strings.TrimSpace(parts[0])
	params := make(map[string]string)
	for _, part := range parts[1:] {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return mimeType, params
}

// parseJSONResponse attempts to parse the JSON error response and update the APIError structure.
func parseJSONResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	if err := json.Unmarshal(bodyBytes, apiError); err != nil {
		apiError.Raw = string(bodyBytes)
		log.LogError("json_parsing_error",
			resp.Request.Method,
			resp.Request.URL.String(),
			resp.StatusCode,
			"JSON parsing failed",
			err,
			apiError.Raw,
		)
	} else {
		// Successfully parsed JSON error, so log the error details.
		logError(log, apiError, "json_error_detected", resp)
	}
}

// parseXMLResponse should be implemented to parse XML responses and log errors using the centralized logger.
func parseXMLResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	var xmlErr APIError

	// Attempt to unmarshal the XML body into the XMLErrorResponse struct
	if err := xml.Unmarshal(bodyBytes, &xmlErr); err != nil {
		// If parsing fails, log the error and keep the raw response
		apiError.Raw = string(bodyBytes)
		log.LogError("xml_parsing_error",
			resp.Request.Method,
			resp.Request.URL.String(),
			apiError.StatusCode,
			fmt.Sprintf("Failed to parse XML: %s", err),
			err,
			apiError.Raw,
		)
	} else {
		// Update the APIError with information from the parsed XML
		apiError.Message = xmlErr.Message
		// Assuming you might want to add a 'Code' field to APIError to store xmlErr.Code
		// apiError.Code = xmlErr.Code

		// Log the parsed error details
		log.LogError("xml_error_detected",
			resp.Request.Method,
			resp.Request.URL.String(),
			apiError.StatusCode,
			"Parsed XML error successfully",
			nil, // No error during parsing
			apiError.Raw,
		)
	}
}

// parseTextResponse updates the APIError structure based on a plain text error response and logs it.
func parseTextResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	bodyText := string(bodyBytes)
	apiError.Message = bodyText
	apiError.Raw = bodyText
	// Log the plain text error using the centralized logger.
	logError(log, apiError, "text_error_detected", resp)
}

// parseHTMLResponse extracts meaningful information from an HTML error response.
func parseHTMLResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	// Always set the Raw field to the entire HTML content for debugging purposes
	apiError.Raw = string(bodyBytes)

	reader := bytes.NewReader(bodyBytes)
	doc, err := html.Parse(reader)
	if err != nil {
		logError(log, apiError, "html_parsing_error", resp)
		return
	}

	var parse func(*html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			if n.FirstChild != nil {
				apiError.Message = n.FirstChild.Data
				// Optionally, you might break or return after finding the first relevant message
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c)
		}
	}

	parse(doc)

	// If no <p> tag was found or it was empty, fallback to using the raw HTML
	if apiError.Message == "" {
		apiError.Message = "HTML Error: See 'Raw' field for details."
		apiError.Raw = string(bodyBytes)
	}
	// Log the extracted error message or the fallback message
	logError(log, apiError, "html_error_detected", resp)
}

// logError logs the error details using the provided logger instance.
func logError(log logger.Logger, apiError *APIError, event string, resp *http.Response) {
	// Prepare the error message. If apiError.Message is empty, use a default message.
	errorMessage := apiError.Message
	if errorMessage == "" {
		errorMessage = "An unspecified error occurred"
	}

	// Use LogError method from the logger package for error logging.
	log.LogError(
		event,
		resp.Request.Method,
		resp.Request.URL.String(),
		apiError.StatusCode,
		resp.Status,
		fmt.Errorf(errorMessage),
		apiError.Raw,
	)
}
