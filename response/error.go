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
	StatusCode  int      `json:"status_code"`       // HTTP status code
	Method      string   `json:"method"`            // HTTP method used for the request
	URL         string   `json:"url"`               // The URL of the HTTP request
	Message     string   `json:"message"`           // Summary of the error
	Details     []string `json:"details,omitempty"` // Detailed error messages, if any
	RawResponse string   `json:"raw_response"`      // Raw response body for debugging
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
	return fmt.Sprintf("API Error: StatusCode=%d, Message=%s", e.StatusCode, e.Message)
}

// HandleAPIErrorResponse handles the HTTP error response from an API and logs the error.
func HandleAPIErrorResponse(resp *http.Response, log logger.Logger) *APIError {
	apiError := &APIError{
		StatusCode: resp.StatusCode,
		Method:     resp.Request.Method,
		URL:        resp.Request.URL.String(),
		Message:    "API Error Response",
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		apiError.RawResponse = "Failed to read response body"
		log.LogError("error_reading_response_body", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, err, apiError.RawResponse)
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
		apiError.RawResponse = string(bodyBytes)
		apiError.Message = "Unknown content type error"
		log.LogError("unknown_content_type_error", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, "Unknown content type", nil, apiError.RawResponse)
	}

	return apiError
}

// parseJSONResponse attempts to parse the JSON error response and update the APIError structure.
func parseJSONResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	if err := json.Unmarshal(bodyBytes, apiError); err != nil {
		apiError.RawResponse = string(bodyBytes)
		log.LogError("json_parsing_error", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, err, apiError.RawResponse)
	} else {
		if apiError.Message == "" {
			apiError.Message = "An unknown error occurred"
		}

		log.LogError("json_error_detected", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, err, apiError.RawResponse)
	}
}

// parseXMLResponse dynamically parses XML error responses and accumulates potential error messages.
func parseXMLResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	// Always set the Raw field to the entire XML content for debugging purposes.
	apiError.RawResponse = string(bodyBytes)

	// Parse the XML document.
	doc, err := xmlquery.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
		log.LogError("xml_parsing_error", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, err, apiError.RawResponse)
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
		logErr = fmt.Errorf("no error message extracted from XML")
	}
	// Log the error or the lack of extracted messages using the centralized logger.
	log.LogError("html_parsing_error", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, logErr, apiError.RawResponse)
}

// parseTextResponse updates the APIError structure based on a plain text error response and logs it.
func parseTextResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	// Convert the body bytes to a string and assign it to both the message and RawResponse fields of APIError.
	bodyText := string(bodyBytes)
	apiError.RawResponse = bodyText

	// Directly use the body text as the error message if the Message field is empty.
	apiError.Message = bodyText

	log.LogError(
		"text_error_detected",     // Event
		resp.Request.Method,       // HTTP method
		resp.Request.URL.String(), // Request URL
		apiError.StatusCode,       // HTTP status code
		resp.Status,               // HTTP status message from the response
		nil,                       // Error (nil because text parsing is unlikely to fail)
		apiError.RawResponse,      // Raw response text
	)
}

// parseHTMLResponse extracts meaningful information from an HTML error response,
// concatenating all text within <p> tags and links found within them.
func parseHTMLResponse(bodyBytes []byte, apiError *APIError, log logger.Logger, resp *http.Response) {
	// Set the entire HTML content as the RawResponse for debugging purposes.
	apiError.RawResponse = string(bodyBytes)

	// Parse the HTML document.
	reader := bytes.NewReader(bodyBytes)
	doc, err := html.Parse(reader)
	if err != nil {
		log.LogError("html_parsing_error", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, err, apiError.RawResponse)
		return
	}

	var messages []string // To accumulate messages and links.
	var parse func(*html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			var pContent strings.Builder
			// Define a function to traverse child nodes of the <p> tag.
			var traverseChildren func(*html.Node)
			traverseChildren = func(c *html.Node) {
				if c.Type == html.TextNode {
					// Append text content directly.
					pContent.WriteString(strings.TrimSpace(c.Data) + " ")
				} else if c.Type == html.ElementNode && c.Data == "a" {
					// Extract href attribute value for links.
					for _, attr := range c.Attr {
						if attr.Key == "href" {
							// Append the link to the pContent builder.
							pContent.WriteString("[Link: " + attr.Val + "] ")
							break
						}
					}
				}
				// Recursively traverse all children of the current node.
				for child := c.FirstChild; child != nil; child = child.NextSibling {
					traverseChildren(child)
				}
			}
			// Start traversing child nodes of the current <p> tag.
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				traverseChildren(child)
			}
			finalContent := strings.TrimSpace(pContent.String())
			if finalContent != "" {
				// Add the content of the <p> tag to messages.
				messages = append(messages, finalContent)
			}
		}
		// Continue traversing the document.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c)
		}
	}

	parse(doc) // Start parsing the document.

	// Concatenate all accumulated messages and links with a separator.
	if len(messages) > 0 {
		apiError.Message = strings.Join(messages, "; ")
	} else {
		// Fallback error message if no specific content was extracted.
		apiError.Message = "HTML Error: See 'Raw' field for details."
	}

	// Determine the error to log based on whether content was found.
	var logErr error
	if apiError.Message == "" {
		logErr = fmt.Errorf("no error message extracted from HTML")
	}

	// Log the extracted content or the fallback message using the centralized logger.
	log.LogError("html_error_detected", resp.Request.Method, resp.Request.URL.String(), apiError.StatusCode, resp.Status, logErr, apiError.RawResponse)
}
