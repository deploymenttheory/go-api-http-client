// graph_api_handler.go
/* ------------------------------Summary----------------------------------------
This is a api handler module for the http_client to accommodate specifics of
microsoft's graph api(s). It handles the encoding (marshalling) and decoding (unmarshalling)
of data. It also sets the correct content headers for the various http methods.

This module integrates with the http_client logger for wrapped error handling
for human readable return codes. It also supports the http_client tiered logging
functionality for logging support.

The logic of this module is defined as follows:
Graph API & Graph Beta API:

For requests (GET, POST, PUT, DELETE):
- Encoding (Marshalling): Use JSON format.
For responses (GET, POST, PUT):
- Decoding (Unmarshalling): Use JSON format.
For responses (DELETE):
- Handle response codes as response body lacks anything useful.
Headers
- Sets accept headers based on weighting.Graph API doesn't support XML, so MIME type is skipped and returns JSON
- Set content header as application/json with edge case exceptions based on need.

*/
package graph

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	_ "embed"

	"github.com/deploymenttheory/go-api-http-client/internal/httpclient"
)

// Endpoint constants represent the URL suffixes used for Graph API token interactions.
const (
	DefaultBaseDomain       = "graph.microsoft.com"           // DefaultBaseDomain: represents the base domain for graph.
	TokenInvalidateEndpoint = "/api/v1/auth/invalidate-token" // TokenInvalidateEndpoint: The endpoint to invalidate an active token.
)

// ConfigMap is a map that associates endpoint URL patterns with their corresponding configurations.
// The map's keys are strings that identify the endpoint, and the values are EndpointConfig structs
// that hold the configuration for that endpoint.
type ConfigMap map[string]EndpointConfig

// Variables
var configMap ConfigMap

// Embedded Resources
//
//go:embed graph_api_exceptions_configuration.json
var graph_api_exceptions_configuration []byte

// Package-level Functions

// init is invoked automatically on package initialization and is responsible for
// setting up the default state of the package by loading the default configuration.
// If an error occurs during the loading process, the program will terminate with a fatal error log.
func init() {
	// Load the default configuration from an embedded resource.
	err := loadDefaultConfig()
	if err != nil {
		log.Fatalf("Error loading default config: %s", err)
	}
}

// loadDefaultConfig reads and unmarshals the graph_api_exceptions_configuration JSON data from an embedded file
// into the configMap variable, which holds the exceptions configuration for endpoint-specific headers.
// Returns an error if the unmarshalling process fails.
func loadDefaultConfig() error {
	// Unmarshal the embedded default configuration into the global configMap.
	return json.Unmarshal(graph_api_exceptions_configuration, &configMap)
}

// LoadUserConfig allows users to apply their own configuration by providing a JSON file.
// The custom configuration will override the default settings previously loaded.
// It reads the file from the provided filename path and unmarshals its content into the configMap.
// If reading or unmarshalling fails, an error is returned.
func LoadUserConfig(filename string) error {
	// Read the user-provided JSON configuration file and unmarshal it into the global configMap.
	userConfigBytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	// Override the default configuration with the user's custom settings.
	return json.Unmarshal(userConfigBytes, &configMap)
}

// Structs

// EndpointConfig is a struct that holds configuration details for a specific API endpoint.
// It includes what type of content it can accept and what content type it should send.
type EndpointConfig struct {
	Accept      string  `json:"accept"`       // Accept specifies the MIME type the endpoint can handle in responses.
	ContentType *string `json:"content_type"` // ContentType, if not nil, specifies the MIME type to set for requests sent to the endpoint. A pointer is used to distinguish between a missing field and an empty string.
}

// UnifiedAPIHandler is a struct that implements the APIHandler interface.
// It holds a Logger instance to facilitate logging across various API handling methods.
// This handler is responsible for encoding and decoding request and response data,
// determining content types, and other API interactions as defined by the APIHandler interface.
type GraphAPIHandler struct {
	logger                       httpclient.Logger // logger is used to output logs for the API handling processes.
	endpointAcceptedFormatsCache map[string][]string
}

// Functions

// ConstructMSGraphAPIEndpoint constructs the full URL for an MS Graph API endpoint.
// The function takes version (e.g., "/v1.0" or "/beta") and the specific API path.
func (g *GraphAPIHandler) ConstructMSGraphAPIEndpoint(endpointPath string) string {
	url := fmt.Sprintf("https://%s%s", DefaultBaseDomain, endpointPath)
	g.logger.Info("Request will be made to MS Graph API URL:", "URL", url)
	return url
}

// GetAPIHandler initializes and returns an APIHandler with a configured logger.
func GetAPIHandler(config Config) APIHandler {
	handler := &GraphAPIHandler{}
	logger := NewDefaultLogger()
	logger.SetLevel(config.LogLevel) // Use the LogLevel from the config
	handler.SetLogger(logger)
	return handler
}

// SetLogger assigns a Logger instance to the UnifiedAPIHandler.
// This allows for logging throughout the handler's operations,
// enabling consistent logging that follows the configuration of the provided Logger.
func (u *GraphAPIHandler) SetLogger(logger Logger) {
	u.logger = logger
}

/*
// GetContentTypeHeader determines the appropriate Content-Type header for a given API endpoint.
// It attempts to find a content type that matches the endpoint prefix in the global configMap.
// If a match is found and the content type is defined (not nil), it returns the specified content type.
// If the content type is nil or no match is found in configMap, it falls back to default behaviors:
// - For all url endpoints it defaults to "application/json" for the graph beta and V1.0 API's.
// If the endpoint does not match any of the predefined patterns, "application/json" is used as a fallback.
// This method logs the decision process at various stages for debugging purposes.
func (u *GraphAPIHandler) GetContentTypeHeader(endpoint string) string {
	// Dynamic lookup from configuration should be the first priority
	for key, config := range configMap {
		if strings.HasPrefix(endpoint, key) {
			if config.ContentType != nil {
				u.logger.Debug("Content-Type for endpoint found in configMap", "endpoint", endpoint, "content_type", *config.ContentType)
				return *config.ContentType
			}
			u.logger.Debug("Content-Type for endpoint is nil in configMap, handling as special case", "endpoint", endpoint)
			// If a nil ContentType is an expected case, do not set Content-Type header.
			return "" // Return empty to indicate no Content-Type should be set.
		}
	}

	// Fallback to JSON if no other match is found.
	u.logger.Debug("Content-Type for endpoint not found in configMap, using default JSON for Graph API for endpoint", endpoint)
	return "application/json"
}
*/

// GetContentTypeHeader determines the appropriate Content-Type header for a given API endpoint.
// It checks a cache of previously fetched accepted formats for the endpoint. If the cache does not
// have the information, it makes an OPTIONS request to fetch and cache these formats. The function
// then selects the most appropriate Content-Type based on the accepted formats, defaulting to
// "application/json" if no specific format is found or in case of an error.
//
// Parameters:
// - endpoint: The API endpoint for which to determine the Content-Type.
//
// Returns:
// - The chosen Content-Type for the request, as a string.
func (u *GraphAPIHandler) GetContentTypeHeader(endpoint string) string {
	// Initialize the cache if it's not already initialized
	if u.endpointAcceptedFormatsCache == nil {
		u.endpointAcceptedFormatsCache = make(map[string][]string)
	}

	// Check the cache first
	if formats, found := u.endpointAcceptedFormatsCache[endpoint]; found {
		u.logger.Debug("Using cached accepted formats", "endpoint", endpoint, "formats", formats)
		for _, format := range formats {
			if format == "application/json" {
				return "application/json"
			}
			if format == "application/xml" {
				return "application/xml"
			}
			if format == "text/html" {
				return "text/html"
			}
			if format == "text/csv" {
				return "text/csv"
			}
			if format == "application/x-www-form-urlencoded" {
				return "application/x-www-form-urlencoded"
			}
			if format == "text/plain" {
				return "text/plain"
			}
			// Additional format conditions can be added here
		}
	} else {
		// Fetch the supported formats as they are not in cache
		formats, err := u.FetchSupportedRequestFormats(endpoint)
		if err != nil {
			u.logger.Warn("Failed to fetch supported request formats from api query, defaulting to 'application/json'", "error", err)
			return "application/json" // Fallback to default
		}

		// Cache the fetched formats
		u.endpointAcceptedFormatsCache[endpoint] = formats
		u.logger.Debug("Fetched and cached accepted formats", "endpoint", endpoint, "formats", formats)

		for _, format := range formats {
			if format == "application/json" {
				return "application/json"
			}
			if format == "application/xml" {
				return "application/xml"
			}
			if format == "text/html" {
				return "text/html"
			}
			if format == "text/csv" {
				return "text/csv"
			}
			if format == "application/x-www-form-urlencoded" {
				return "application/x-www-form-urlencoded"
			}
			if format == "text/plain" {
				return "text/plain"
			}
		}
	}

	return "application/json" // Default to JSON if no suitable format is found
}

// FetchSupportedRequestFormats sends an OPTIONS request to the specified API endpoint
// and parses the response to extract the MIME types that the endpoint can accept for requests.
// This function is useful for dynamically determining the supported formats (like JSON, XML, etc.)
// for an endpoint, which can then be used to set the appropriate 'Content-Type' header in subsequent requests.
//
// Parameters:
// - endpoint: A string representing the API endpoint for which to fetch the supported request formats.
//
// Returns:
//   - A slice of strings, where each string is a MIME type that the endpoint can accept.
//     Example: []string{"application/json", "application/xml"}
//   - An error if the request could not be sent, the response could not be processed, or if the endpoint
//     does not specify accepted formats in its response headers.
//
// Note:
//   - The function makes an HTTP OPTIONS request to the given endpoint and reads the 'Accept' header in the response.
//   - If the 'Accept' header is not present or the OPTIONS method is not supported by the endpoint, the function
//     returns an error.
//   - It is the responsibility of the caller to handle any errors and to decide on the default action
//     if no formats are returned or in case of an error.
func (u *GraphAPIHandler) FetchSupportedRequestFormats(endpoint string) ([]string, error) {
	url := fmt.Sprintf("https://%s%s", DefaultBaseDomain, endpoint)
	req, err := http.NewRequest(http.MethodOptions, url, nil)
	if err != nil {
		return nil, err
	}

	// Add necessary headers, authentication etc.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the Accept header
	acceptHeader := resp.Header.Get("Accept")
	if acceptHeader == "" {
		return nil, fmt.Errorf("no Accept header present in api response")
	}

	formats := strings.Split(acceptHeader, ",")
	return formats, nil
}

// MarshalRequest encodes the request body according to the endpoint for the API.
func (u *GraphAPIHandler) MarshalRequest(body interface{}, method string, endpoint string) ([]byte, error) {
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
			u.logger.Trace("XML Request Body:", "Body", string(data))
		}

	case "json":
		data, err = json.Marshal(body)
		if err != nil {
			u.logger.Error("Failed marshaling JSON request", "error", err)
			return nil, err
		}

		if method == "POST" || method == "PUT" || method == "PATCH" {
			u.logger.Debug("JSON Request Body:", string(data))
		}
	}

	return data, nil
}

// UnmarshalResponse decodes the response body from XML or JSON format depending on the Content-Type header.
func (u *GraphAPIHandler) UnmarshalResponse(resp *http.Response, out interface{}) error {
	// Handle DELETE method
	if resp.Request.Method == "DELETE" {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		} else {
			return fmt.Errorf("DELETE request failed with status code: %d", resp.StatusCode)
		}
	}

	// Handle PATCH method
	if resp.Request.Method == "PATCH" {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		} else {
			return fmt.Errorf("PATCH request failed with status code: %d", resp.StatusCode)
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		u.logger.Error("Failed reading response body", "error", err)
		return err
	}

	// Log the raw response body and headers
	u.logger.Trace("Raw HTTP Response:", string(bodyBytes))
	u.logger.Debug("Unmarshaling response", "status", resp.Status)

	// Log headers when in debug mode
	u.logger.Debug("HTTP Response Headers:", resp.Header)

	// Check the Content-Type and Content-Disposition headers
	contentType := resp.Header.Get("Content-Type")
	contentDisposition := resp.Header.Get("Content-Disposition")

	// Handle binary data if necessary
	if err := u.handleBinaryData(contentType, contentDisposition, bodyBytes, out); err != nil {
		return err
	}

	// If content type is HTML, extract the error message
	if strings.Contains(contentType, "text/html") {
		errMsg := extractErrorMessageFromHTML(string(bodyBytes))
		u.logger.Warn("Received HTML content", "error_message", errMsg, "status_code", resp.StatusCode)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    errMsg,
		}
	}

	// Check for non-success status codes before attempting to unmarshal
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse the error details from the response body for JSON content type
		if strings.Contains(contentType, "application/json") {
			var structuredErr StructuredError
			if jsonErr := json.Unmarshal(bodyBytes, &structuredErr); jsonErr == nil {
				detailedMessage := fmt.Sprintf("%s: %s", structuredErr.Error.Code, structuredErr.Error.Message)
				u.logger.Error("Received API error response", "status_code", resp.StatusCode, "error", detailedMessage)
				return &APIError{
					StatusCode: resp.StatusCode,
					Message:    detailedMessage,
				}
			} else {
				u.logger.Error("Failed to parse JSON error response", "error", jsonErr)
				return fmt.Errorf("received non-success status code: %d and failed to parse error response", resp.StatusCode)
			}
		}

		// If the response is not JSON or another error occurs, return a generic error message
		u.logger.Error("Received non-success status code", "status_code", resp.StatusCode)
		return fmt.Errorf("received non-success status code: %d", resp.StatusCode)
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
			errMsg := extractErrorMessageFromHTML(string(bodyBytes))
			u.logger.Warn("Received HTML content instead of expected format", "error_message", errMsg, "status_code", resp.StatusCode)
			return fmt.Errorf(errMsg)
		}

		// Log the error and return it
		u.logger.Error("Failed to unmarshal response", "error", err)
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return nil
}

// GetAcceptHeader constructs and returns a weighted Accept header string for HTTP requests.
// The Accept header indicates the MIME types that the client can process and prioritizes them
// based on the quality factor (q) parameter. Higher q-values signal greater preference.
// This function specifies a range of MIME types with their respective weights, ensuring that
// the server is informed of the client's versatile content handling capabilities while
// indicating a preference for XML. The specified MIME types cover common content formats like
// images, JSON, XML, HTML, plain text, and certificates, with a fallback option for all other types.
func (u *GraphAPIHandler) GetAcceptHeader() string {
	weightedAcceptHeader := "application/x-x509-ca-cert;q=0.95," +
		"application/pkix-cert;q=0.94," +
		"application/pem-certificate-chain;q=0.93," +
		"application/octet-stream;q=0.8," + // For general binary files
		"image/png;q=0.75," +
		"image/jpeg;q=0.74," +
		"image/*;q=0.7," +
		"application/xml;q=0.65," +
		"text/xml;q=0.64," +
		"text/xml;charset=UTF-8;q=0.63," +
		"application/json;q=0.5," +
		"text/html;q=0.5," +
		"text/plain;q=0.4," +
		"*/*;q=0.05" // Fallback for any other types
	return weightedAcceptHeader
}

// MarshalMultipartFormData takes a map with form fields and file paths and returns the encoded body and content type.
func (u *GraphAPIHandler) MarshalMultipartRequest(fields map[string]string, files map[string]string) ([]byte, string, error) {
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
func (u *GraphAPIHandler) handleBinaryData(contentType, contentDisposition string, bodyBytes []byte, out interface{}) error {
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
