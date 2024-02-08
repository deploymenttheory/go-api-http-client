// jamfpro_api_handler.go
/* ------------------------------Summary----------------------------------------
This is a api handler module for the http_client to accommodate specifics of
jamf's api(s). It handles the encoding (marshalling) and decoding (unmarshalling)
of data. It also sets the correct content headers for the various http methods.

This module integrates with the http_client logger for wrapped error handling
for human readable return codes. It also supports the http_client tiered logging
functionality for logging support.

The logic of this module is defined as follows:
Classic API:

For requests (GET, POST, PUT, DELETE):
- Encoding (Marshalling): Use XML format.
For responses (GET, POST, PUT):
- Decoding (Unmarshalling): Use XML format.
For responses (DELETE):
- Handle response codes as response body lacks anything useful.
Headers
- Sets accept headers based on weighting. XML out weighs JSON to ensure XML is returned
- Sets content header as application/xml with edge case exceptions based on need.

JamfPro API:

For requests (GET, POST, PUT, DELETE):
- Encoding (Marshalling): Use JSON format.
For responses (GET, POST, PUT):
- Decoding (Unmarshalling): Use JSON format.
For responses (DELETE):
- Handle response codes as response body lacks anything useful.
Headers
- Sets accept headers based on weighting. Jamf Pro API doesn't support XML, so MIME type is skipped and returns JSON
- Set content header as application/json with edge case exceptions based on need.
*/
package jamfpro

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

	"github.com/deploymenttheory/go-api-http-client/errors"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// Endpoint constants represent the URL suffixes used for Jamf API token interactions.
const (
	DefaultBaseDomain       = ".jamfcloud.com"                // DefaultBaseDomain: represents the base domain for the jamf instance.
	OAuthTokenEndpoint      = "/api/oauth/token"              // OAuthTokenEndpoint: The endpoint to obtain an OAuth token.
	BearerTokenEndpoint     = "/api/v1/auth/token"            // BearerTokenEndpoint: The endpoint to obtain a bearer token.
	TokenRefreshEndpoint    = "/api/v1/auth/keep-alive"       // TokenRefreshEndpoint: The endpoint to refresh an existing token.
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
//go:embed jamfpro_api_exceptions_configuration.json
var jamfpro_api_exceptions_configuration []byte

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

// loadDefaultConfig reads and unmarshals the jamfpro_api_exceptions_configuration JSON data from an embedded file
// into the configMap variable, which holds the exceptions configuration for endpoint-specific headers.
// Returns an error if the unmarshalling process fails.
func loadDefaultConfig() error {
	// Unmarshal the embedded default configuration into the global configMap.
	return json.Unmarshal(jamfpro_api_exceptions_configuration, &configMap)
}

// EndpointConfig is a struct that holds configuration details for a specific API endpoint.
// It includes what type of content it can accept and what content type it should send.
type EndpointConfig struct {
	Accept      string  `json:"accept"`       // Accept specifies the MIME type the endpoint can handle in responses.
	ContentType *string `json:"content_type"` // ContentType, if not nil, specifies the MIME type to set for requests sent to the endpoint. A pointer is used to distinguish between a missing field and an empty string.
}

// JamfAPIHandler implements the APIHandler interface for the Jamf Pro API.
type JamfAPIHandler struct {
	OverrideBaseDomain string // OverrideBaseDomain is used to override the base domain for URL construction.
	InstanceName       string // InstanceName is the name of the Jamf instance.
}

type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// Functions

func (j *JamfAPIHandler) GetDefaultBaseDomain() string {
	return DefaultBaseDomain
}

func (j *JamfAPIHandler) GetOAuthTokenEndpoint() string {
	return OAuthTokenEndpoint
}

func (j *JamfAPIHandler) GetBearerTokenEndpoint() string {
	return BearerTokenEndpoint
}

func (j *JamfAPIHandler) GetTokenRefreshEndpoint() string {
	return TokenRefreshEndpoint
}

func (j *JamfAPIHandler) GetTokenInvalidateEndpoint() string {
	return TokenInvalidateEndpoint
}

// GetBaseDomain returns the appropriate base domain for URL construction.
// It uses OverrideBaseDomain if set, otherwise falls back to DefaultBaseDomain.
func (j *JamfAPIHandler) GetBaseDomain() string {
	if j.OverrideBaseDomain != "" {
		return j.OverrideBaseDomain
	}
	return DefaultBaseDomain
}

// ConstructAPIResourceEndpoint constructs the full URL for a Jamf API resource endpoint path and logs the URL.
func (j *JamfAPIHandler) ConstructAPIResourceEndpoint(endpointPath string, log logger.Logger) string {
	baseDomain := j.GetBaseDomain()
	url := fmt.Sprintf("https://%s%s%s", j.InstanceName, baseDomain, endpointPath)
	log.Info("Constructed API resource endpoint URL", zap.String("URL", url))
	return url
}

// ConstructAPIAuthEndpoint constructs the full URL for a Jamf API auth endpoint path and logs the URL.
func (j *JamfAPIHandler) ConstructAPIAuthEndpoint(endpointPath string, log logger.Logger) string {
	baseDomain := j.GetBaseDomain()
	url := fmt.Sprintf("https://%s%s%s", j.InstanceName, baseDomain, endpointPath)
	log.Info("Constructed API authentication URL", zap.String("URL", url))
	return url
}

// GetContentTypeHeader determines the appropriate Content-Type header for a given API endpoint.
// It attempts to find a content type that matches the endpoint prefix in the global configMap.
// If a match is found and the content type is defined (not nil), it returns the specified content type.
// If the content type is nil or no match is found in configMap, it falls back to default behaviors:
// - For url endpoints starting with "/JSSResource", it defaults to "application/xml" for the Classic API.
// - For url endpoints starting with "/api", it defaults to "application/json" for the JamfPro API.
// If the endpoint does not match any of the predefined patterns, "application/json" is used as a fallback.
// This method logs the decision process at various stages for debugging purposes.
func (u *JamfAPIHandler) GetContentTypeHeader(endpoint string, log logger.Logger) string {
	// Dynamic lookup from configuration should be the first priority
	for key, config := range configMap {
		if strings.HasPrefix(endpoint, key) {
			if config.ContentType != nil {
				log.Debug("Content-Type for endpoint found in configMap", zap.String("endpoint", endpoint), zap.String("content_type", *config.ContentType))
				return *config.ContentType
			}
			log.Debug("Content-Type for endpoint is nil in configMap, handling as special case", zap.String("endpoint", endpoint))
			// If a nil ContentType is an expected case, do not set Content-Type header.
			return "" // Return empty to indicate no Content-Type should be set.
		}
	}

	// If no specific configuration is found, then check for standard URL patterns.
	if strings.Contains(endpoint, "/JSSResource") {
		log.Debug("Content-Type for endpoint defaulting to XML for Classic API", zap.String("endpoint", endpoint))
		return "application/xml" // Classic API uses XML
	} else if strings.Contains(endpoint, "/api") {
		log.Debug("Content-Type for endpoint defaulting to JSON for JamfPro API", zap.String("endpoint", endpoint))
		return "application/json" // JamfPro API uses JSON
	}

	// Fallback to JSON if no other match is found.
	log.Debug("Content-Type for endpoint not found in configMap or standard patterns, using default JSON", zap.String("endpoint", endpoint))
	return "application/json"
}

// MarshalRequest encodes the request body according to the endpoint for the API.
func (u *JamfAPIHandler) MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error) {
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
			log.Debug("XML Request Body", zap.String("Body", string(data)))
		}

	case "json":
		data, err = json.Marshal(body)
		if err != nil {
			log.Error("Failed marshaling JSON request", zap.Error(err))
			return nil, err
		}

		if method == "POST" || method == "PUT" || method == "PATCH" {
			log.Debug("JSON Request Body", zap.String("Body", string(data)))
		}
	}

	return data, nil
}

// UnmarshalResponse decodes the response body from XML or JSON format depending on the Content-Type header.
func (u *JamfAPIHandler) UnmarshalResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	// Handle DELETE method
	if resp.Request.Method == "DELETE" {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		} else {
			return log.Error("DELETE request failed", zap.Int("Status Code", resp.StatusCode))
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed reading response body", zap.Error(err))
		return err
	}

	// Log the raw response body and headers
	log.Debug("Raw HTTP Response", zap.String("Body", string(bodyBytes)))
	log.Debug("Unmarshaling response", zap.String("status", resp.Status))

	// Log headers when in debug mode
	log.Debug("HTTP Response Headers", zap.Any("Headers", resp.Header))

	// Check the Content-Type and Content-Disposition headers
	contentType := resp.Header.Get("Content-Type")
	contentDisposition := resp.Header.Get("Content-Disposition")

	// Handle binary data if necessary
	if err := u.handleBinaryData(contentType, contentDisposition, bodyBytes, out); err != nil {
		return err
	}

	// If content type is HTML, extract the error message
	if strings.Contains(contentType, "text/html") {
		errMsg := ExtractErrorMessageFromHTML(string(bodyBytes))
		log.Warn("Received HTML content", zap.String("error_message", errMsg), zap.Int("status_code", resp.StatusCode))
		return &errors.APIError{
			StatusCode: resp.StatusCode,
			Message:    errMsg,
		}
	}

	// Check for non-success status codes before attempting to unmarshal
	// Check for non-success status codes before attempting to unmarshal
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse the error details from the response body for JSON content type
		if strings.Contains(contentType, "application/json") {
			description, err := ParseJSONErrorResponse(bodyBytes)
			if err != nil {
				// Log the error using the structured logger and return the error
				log.Error("Failed to parse JSON error response",
					zap.Error(err),
					zap.Int("status_code", resp.StatusCode),
				)
				return err
			}
			// Log the error with description using the structured logger and return the error
			log.Error("Received non-success status code with JSON response",
				zap.Int("status_code", resp.StatusCode),
				zap.String("error_description", description),
			)
			return fmt.Errorf("received non-success status code with JSON response: %s", description)
		}

		// If the response is not JSON or another error occurs, log a generic error message and return an error
		log.Error("Received non-success status code without JSON response",
			zap.Int("status_code", resp.StatusCode),
		)
		return fmt.Errorf("received non-success status code without JSON response: %d", resp.StatusCode)
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
			errMsg := ExtractErrorMessageFromHTML(string(bodyBytes))

			// Log the warning and return an error using the structured logger
			log.Warn("Received HTML content instead of expected format",
				zap.String("error_message", errMsg),
				zap.Int("status_code", resp.StatusCode),
			)
			return fmt.Errorf("received HTML content instead of expected format: %s", errMsg)
		}
	}
	return err
}

// GetAcceptHeader constructs and returns a weighted Accept header string for HTTP requests.
// The Accept header indicates the MIME types that the client can process and prioritizes them
// based on the quality factor (q) parameter. Higher q-values signal greater preference.
// This function specifies a range of MIME types with their respective weights, ensuring that
// the server is informed of the client's versatile content handling capabilities while
// indicating a preference for XML. The specified MIME types cover common content formats like
// images, JSON, XML, HTML, plain text, and certificates, with a fallback option for all other types.
func (u *JamfAPIHandler) GetAcceptHeader() string { // Add closing parenthesis after the function signature
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
func (u *JamfAPIHandler) MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, error) {
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
func (u *JamfAPIHandler) handleBinaryData(contentType, contentDisposition string, bodyBytes []byte, out interface{}) error {
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
