// response/error.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"golang.org/x/net/html"
)

// APIError represents an api error response.
type APIError struct {
	StatusCode int                    `json:"status_code"`      // HTTP status code
	Type       string                 `json:"type"`             // Type of error
	Message    string                 `json:"message"`          // Human-readable message
	Detail     string                 `json:"detail,omitempty"` // Detailed error message
	Errors     map[string]interface{} `json:"errors,omitempty"` // Additional error details
	Raw        string                 `json:"raw"`              // Raw response body for debugging
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
		Type:       "API Error Response",
		Message:    "An error occurred",
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		apiError.Raw = "Failed to read response body"
		log.LogError("error_reading_response_body", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, err, apiError.Raw)
		return apiError
	}

	mimeType, _ := ParseContentTypeHeader(resp.Header.Get("Content-Type"))
	switch mimeType {
	case "application/json":
		parseJSONResponse(bodyBytes, apiError, log, resp)
	case "application/xml", "text/xml":
		parseXMLResponse(bodyBytes, apiError, log, resp)
	case "text/html":
		parseHTMLResponse(bodyBytes, apiError, log, resp)
	case "text/plain":
		parseTextResponse(bodyBytes, apiError, log, resp)
	default:
		apiError.Raw = string(bodyBytes)
		apiError.Message = "Unknown content type error"
		log.LogError("unknown_content_type_error", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, "Unknown content type", nil, apiError.Raw)
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
		logError(log, apiError, "json_parsing_error", resp.Request.Method, resp.Request.URL.String(), resp.Status, err)
	} else {
		if apiError.Message == "" {
			apiError.Message = "An unknown error occurred"
		}

		// Log the detected JSON error with all the context information.
		logError(log, apiError, "json_error_detected", resp.Request.Method, resp.Request.URL.String(), resp.Status, nil)
	}
}

// parseXMLResponse dynamically parses XML error responses and accumulates potential error messages.
func parseXMLResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	// Always set the Raw field to the entire XML content for debugging purposes.
	apiError.Raw = string(bodyBytes)

	// Parse the XML document.
	doc, err := xmlquery.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
		// Log the XML parsing error with all the context information.
		logError(log, apiError, "xml_parsing_error", resp.Request.Method, resp.Request.URL.String(), resp.Status, err)
		return
	}

	var messages []string
	var traverse func(*xmlquery.Node)
	traverse = func(n *xmlquery.Node) {
		if n.Type == xmlquery.TextNode && strings.TrimSpace(n.Data) != "" {
			messages = append(messages, strings.TrimSpace(n.Data))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	// Concatenate all messages found in the XML for the 'Message' field of APIError.
	if len(messages) > 0 {
		apiError.Message = strings.Join(messages, "; ")
	} else {
		apiError.Message = "Failed to extract error details from XML response"
	}

	// Determine the error to log based on whether a message was found.
	var logErr error
	if apiError.Message == "" {
		logErr = fmt.Errorf("No error message extracted from XML")
	}

	// Log the error or the lack of extracted messages using the centralized logger.
	logError(log, apiError, "xml_error_detected", resp.Request.Method, resp.Request.URL.String(), resp.Status, logErr)
}

// parseTextResponse updates the APIError structure based on a plain text error response and logs it.
func parseTextResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	bodyText := string(bodyBytes)
	apiError.Raw = bodyText

	// Check if the 'Message' field of APIError is empty and use the body text as the message.
	if apiError.Message == "" {
		apiError.Message = bodyText
	}

	// Use the updated logError function with the additional parameters.
	logError(log, apiError, "text_error_detected", resp.Request.Method, resp.Request.URL.String(), resp.Status, nil)
}

// parseHTMLResponse extracts meaningful information from an HTML error response and concatenates all text within <p> tags.
func parseHTMLResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	// Always set the Raw field to the entire HTML content for debugging purposes.
	apiError.Raw = string(bodyBytes)

	reader := bytes.NewReader(bodyBytes)
	doc, err := html.Parse(reader)
	if err != nil {
		// Log HTML parsing error using centralized logger with context.
		logError(log, apiError, "html_parsing_error", resp.Request.Method, resp.Request.URL.String(), resp.Status, err)
		return
	}

	var messages []string // To accumulate messages from all <p> tags.
	var parse func(*html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			var pText strings.Builder
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
					// Build text content of <p> tag.
					if pText.Len() > 0 {
						pText.WriteString(" ") // Add a space between text nodes within the same <p> tag.
					}
					pText.WriteString(strings.TrimSpace(c.Data))
				}
			}
			if pText.Len() > 0 {
				// Add the built text content of the <p> tag to messages.
				messages = append(messages, pText.String())
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c) // Recursively parse the document.
		}
	}

	parse(doc)

	// Concatenate all accumulated messages with a separator.
	if len(messages) > 0 {
		apiError.Message = strings.Join(messages, "; ")
	} else {
		// Fallback error message if no specific messages were extracted.
		apiError.Message = "HTML Error: See 'Raw' field for details."
	}

	// Determine the error to log based on whether a message was found.
	var logErr error
	if apiError.Message == "" {
		logErr = fmt.Errorf("No error message extracted from HTML")
	}

	// Log the extracted error message or the fallback message using the centralized logger.
	logError(log, apiError, "html_error_detected", resp.Request.Method, resp.Request.URL.String(), resp.Status, logErr)
}

// logError logs the error details using the provided logger instance.
// func logError(log logger.Logger, apiError *APIError, event string, resp *http.Response) {
// 	// Prepare the error message. If apiError.Message is empty, use a default message.
// 	errorMessage := apiError.Message
// 	if errorMessage == "" {
// 		errorMessage = "An unspecified error occurred"
// 	}

// 	// Use LogError method from the logger package for error logging.
// 	log.LogError(
// 		event,
// 		resp.Request.Method,
// 		resp.Request.URL.String(),
// 		apiError.StatusCode,
// 		resp.Status,
// 		fmt.Errorf(errorMessage),
// 		apiError.Raw,
// 	)
// }

func logError(log logger.Logger, apiError *APIError, event, method, url, statusMessage string, err error) {
	// Prepare the error message. If apiError.Message is empty, use a default message.
	errorMessage := apiError.Message
	if errorMessage == "" {
		errorMessage = "An unspecified error occurred"
	}

	// Call the LogError method from the logger package for error logging.
	log.LogError(event, method, url, apiError.StatusCode, statusMessage, err, apiError.Raw)
}
