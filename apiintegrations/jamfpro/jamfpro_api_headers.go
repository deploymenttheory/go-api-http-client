// jamfpro_api_headers.go
package jamfpro

import (
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// GetContentTypeHeader determines the appropriate Content-Type header for a given API endpoint.
// It attempts to find a content type that matches the endpoint prefix in the global configMap.
// If a match is found and the content type is defined (not nil), it returns the specified content type.
// If the content type is nil or no match is found in configMap, it falls back to default behaviors:
// - For url endpoints starting with "/JSSResource", it defaults to "application/xml" for the Classic API.
// - For url endpoints starting with "/api", it defaults to "application/json" for the JamfPro API.
// If the endpoint does not match any of the predefined patterns, "application/json" is used as a fallback.
// This method logs the decision process at various stages for debugging purposes.
func (j *JamfAPIHandler) GetContentTypeHeader(endpoint string, log logger.Logger) string {
	// Dynamic lookup from configuration should be the first priority
	for key, config := range configMap {
		if strings.HasPrefix(endpoint, key) {
			if config.ContentType != nil {
				j.Logger.Debug("Content-Type for endpoint found in configMap", zap.String("endpoint", endpoint), zap.String("content_type", *config.ContentType))
				return *config.ContentType
			}
			j.Logger.Debug("Content-Type for endpoint is nil in configMap, handling as special case", zap.String("endpoint", endpoint))
			// If a nil ContentType is an expected case, do not set Content-Type header.
			return "" // Return empty to indicate no Content-Type should be set.
		}
	}

	// Special case for package upload endpoint
	if strings.HasPrefix(endpoint, "/api/v1/packages") && strings.HasSuffix(endpoint, "/upload") {
		j.Logger.Debug("Skipping Content-Type setting for package upload endpoint. Multipart request will handling setting directly with boundary", zap.String("endpoint", endpoint))
		return "" // Skip setting Content-Type here
	}

	// If no specific configuration is found, then check for standard URL patterns.
	if strings.Contains(endpoint, "/JSSResource") {
		j.Logger.Debug("Content-Type for endpoint defaulting to XML for Classic API", zap.String("endpoint", endpoint))
		return "application/xml" // Classic API uses XML
	} else if strings.Contains(endpoint, "/api") {
		j.Logger.Debug("Content-Type for endpoint defaulting to JSON for JamfPro API", zap.String("endpoint", endpoint))
		return "application/json" // JamfPro API uses JSON
	}

	// Fallback to JSON if no other match is found.
	j.Logger.Debug("Content-Type for endpoint not found in configMap or standard patterns, using default JSON", zap.String("endpoint", endpoint))
	return "application/json"
}

// GetAcceptHeader constructs and returns a weighted Accept header string for HTTP requests.
// The Accept header indicates the MIME types that the client can process and prioritizes them
// based on the quality factor (q) parameter. Higher q-values signal greater preference.
// This function specifies a range of MIME types with their respective weights, ensuring that
// the server is informed of the client's versatile content handling capabilities while
// indicating a preference for XML. The specified MIME types cover common content formats like
// images, JSON, XML, HTML, plain text, and certificates, with a fallback option for all other types.
func (j *JamfAPIHandler) GetAcceptHeader() string {
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

// GetAPIRequestHeaders returns a map of standard headers required for making API requests.
func (j *JamfAPIHandler) GetAPIRequestHeaders(endpoint string) map[string]string {
	headers := map[string]string{
		"Accept":        j.GetAcceptHeader(),                        // Dynamically set based on API requirements.
		"Content-Type":  j.GetContentTypeHeader(endpoint, j.Logger), // Dynamically set based on the endpoint.
		"Authorization": "",                                         // To be set by the client with the actual token.
		"User-Agent":    "go-api-http-client-jamfpro-handler",       // To be set by the client, usually with application info.
	}
	return headers
}
