// api_handler.go
package httpclient

import (
	"fmt"
	"net/http"

	"github.com/deploymenttheory/go-api-http-client/internal/apihandlers/jamfpro"
)

// APIHandler is an interface for encoding, decoding, and determining content types for different API implementations.
// It encapsulates behavior for encoding and decoding requests and responses.
type APIHandler interface {
	GetBaseDomain() string
	ConstructAPIResourceEndpoint(endpointPath string) string
	ConstructAPIAuthEndpoint(endpointPath string, logger Logger) string
	MarshalRequest(body interface{}, method string, endpoint string, logger Logger) ([]byte, error)
	MarshalMultipartRequest(fields map[string]string, files map[string]string, logger Logger) ([]byte, string, error)
	UnmarshalResponse(resp *http.Response, out interface{}, logger Logger) error
	GetContentTypeHeader(method string, logger Logger) string
	GetAcceptHeader(logger Logger) string
}

// LoadAPIHandler returns an APIHandler based on the provided API type.
// 'apiType' parameter could be "jamf" or "graph" to specify which API handler to load.
func LoadAPIHandler(config Config, apiType string) (APIHandler, error) {
	var apiHandler APIHandler
	switch apiType {
	case "jamfpro":
		// Assuming GetAPIHandler returns a JamfAPIHandler
		apiHandler = &jamfpro.JamfAPIHandler{
			// Initialize with necessary parameters
		}
	/*case "graph":
	// Assuming GetAPIHandler returns a GraphAPIHandler
	apiHandler = &graph.GraphAPIHandler{
		// Initialize with necessary parameters
	}*/
	default:
		return nil, fmt.Errorf("unsupported API type: %s", apiType)
	}

	// Set the logger level for the handler if needed
	logger := NewDefaultLogger() // Or use config.Logger if it's not nil
	logger.SetLevel(config.LogLevel)

	return apiHandler, nil
}
