// graph_api_error_messages.go
package graph

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// APIHandlerError represents an error response from the graph API.
type APIHandlerError struct {
	HTTPStatusCode int                    `json:"httpStatusCode"`
	ErrorType      string                 `json:"errorType"`
	ErrorMessage   string                 `json:"errorMessage"`
	ExtraDetails   map[string]interface{} `json:"extraDetails"`
}

// ReturnAPIErrorResponse parses an HTTP error response from the graph API.
func (g *GraphAPIHandler) ReturnAPIErrorResponse(resp *http.Response) *APIHandlerError {
	var errorMessage, errorType string
	var extraDetails map[string]interface{}

	// Safely read the response body
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return &APIHandlerError{
			HTTPStatusCode: resp.StatusCode,
			ErrorType:      "ReadError",
			ErrorMessage:   "Failed to read response body",
		}
	}

	// Ensure the body can be re-read for subsequent operations
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	contentType := resp.Header.Get("Content-Type")

	// Handle JSON content type
	if strings.Contains(contentType, "application/json") {
		description, parseErr := ParseJSONErrorResponse(bodyBytes)
		if parseErr == nil {
			errorMessage = description
			errorType = "JSONError"
		} else {
			errorMessage = "Failed to parse JSON error response: " + parseErr.Error()
		}
	} else if strings.Contains(contentType, "text/html") {
		// Handle HTML content type
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			errorMessage = ExtractErrorMessageFromHTML(string(bodyBytes))
			errorType = "HTMLError"
		} else {
			errorMessage = "Failed to read response body for HTML error parsing"
		}
	} else {
		// Fallback for unhandled content types
		errorMessage = "An unknown error occurred"
	}

	return &APIHandlerError{
		HTTPStatusCode: resp.StatusCode,
		ErrorType:      errorType,
		ErrorMessage:   errorMessage,
		ExtraDetails:   extraDetails,
	}
}

// ExtractErrorMessageFromHTML attempts to parse an HTML error page and extract a combined human-readable error message.
func ExtractErrorMessageFromHTML(htmlContent string) string {
	r := bytes.NewReader([]byte(htmlContent))
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "Unable to parse HTML content"
	}

	var messages []string
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			messages = append(messages, text)
		}
	})

	combinedMessage := strings.Join(messages, " - ")
	return combinedMessage
}

// ParseJSONErrorResponse parses the JSON error message from the response body.
func ParseJSONErrorResponse(body []byte) (string, error) {
	var errorResponse struct {
		HTTPStatus int `json:"httpStatus"`
		Errors     []struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			ID          string `json:"id"`
			Field       string `json:"field"`
		} `json:"errors"`
	}

	err := json.Unmarshal(body, &errorResponse)
	if err != nil {
		return "", err
	}

	if len(errorResponse.Errors) > 0 {
		return errorResponse.Errors[0].Description, nil
	}

	return "No error description available", nil
}
