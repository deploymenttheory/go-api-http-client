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
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

// APIError represents an api error response.
type APIError struct {
	StatusCode  int      `json:"status_code"` // HTTP status code
	Method      string   `json:"method"`      // HTTP method used for the request
	URL         string   `json:"url"`         // The URL of the HTTP request
	HTTPStatus  int      `json:"httpStatus,omitempty"`
	Errors      []Errors `json:"errors,omitempty"`
	Message     string   `json:"message"`           // Summary of the error
	Details     []string `json:"details,omitempty"` // Detailed error messages, if any
	RawResponse string   `json:"raw_response"`      // Raw response body for debugging
}

// Errors represents individual error details within an API error response.
type Errors struct {
	Code        string  `json:"code,omitempty"`
	Field       string  `json:"field,omitempty"`
	Description string  `json:"description,omitempty"`
	ID          *string `json:"id,omitempty"`
}

// Error returns a string representation of the APIError, making it compatible with the error interface.
func (e *APIError) Error() string {
	data, err := json.Marshal(e)
	if err == nil {
		return string(data)
	}

	if e.Message == "" {
		e.Message = http.StatusText(e.StatusCode)
	}

	return fmt.Sprintf("API Error: StatusCode=%d, Message=%s", e.StatusCode, e.Message)
}

// HandleAPIErrorResponse handles the HTTP error response from an API and logs the error.
func HandleAPIErrorResponse(resp *http.Response, sugar *zap.SugaredLogger) *APIError {
	apiError := &APIError{
		StatusCode: resp.StatusCode,
		Method:     resp.Request.Method,
		URL:        resp.Request.URL.String(),
		Message:    "API Error Response",
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		apiError.RawResponse = "Failed to read response body"
		return apiError
	}

	mimeType, _ := parseHeader(resp.Header.Get("Content-Type"))
	switch mimeType {
	case "application/json":
		parseJSONResponse(bodyBytes, apiError)
	case "application/xml", "text/xml":
		parseXMLResponse(bodyBytes, apiError)
	case "text/html":
		parseHTMLResponse(bodyBytes, apiError)
	case "text/plain":
		parseTextResponse(bodyBytes, apiError)
	default:
		apiError.RawResponse = string(bodyBytes)
		apiError.Message = "Unknown content type error"
	}

	return apiError
}

// parseJSONResponse attempts to parse the JSON error response and update the APIError structure.
func parseJSONResponse(bodyBytes []byte, apiError *APIError) {
	if err := json.Unmarshal(bodyBytes, apiError); err != nil {
		apiError.RawResponse = string(bodyBytes)
	} else {
		if apiError.Message == "" {
			apiError.Message = "An unknown error occurred"
		}

	}
}

// parseXMLResponse dynamically parses XML error responses and accumulates potential error messages.
func parseXMLResponse(bodyBytes []byte, apiError *APIError) {
	apiError.RawResponse = string(bodyBytes)

	doc, err := xmlquery.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
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

	if len(messages) > 0 {
		apiError.Message = strings.Join(messages, "; ")
	} else {
		apiError.Message = "Failed to extract error details from XML response"
	}

}

// parseTextResponse updates the APIError structure based on a plain text error response and logs it.
func parseTextResponse(bodyBytes []byte, apiError *APIError) {
	bodyText := string(bodyBytes)
	apiError.RawResponse = bodyText
	apiError.Message = bodyText
}

// parseHTMLResponse extracts meaningful information from an HTML error response,
// concatenating all text within <p> tags and links found within them.
func parseHTMLResponse(bodyBytes []byte, apiError *APIError) {
	apiError.RawResponse = string(bodyBytes)

	reader := bytes.NewReader(bodyBytes)
	doc, err := html.Parse(reader)
	if err != nil {
		return
	}

	var messages []string
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

	parse(doc)

	if len(messages) > 0 {
		apiError.Message = strings.Join(messages, "; ")
	} else {
		apiError.Message = "HTML Error: See 'Raw' field for details."
	}

}
