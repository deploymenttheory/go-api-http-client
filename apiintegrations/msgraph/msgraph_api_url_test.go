// apiintegrations/msgraph/msgraph_api_url_test.go
package msgraph

import (
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestConstructAPIResourceEndpoint tests the ConstructAPIResourceEndpoint function.
func TestConstructAPIResourceEndpoint(t *testing.T) {
	const baseURL = "https://graph.microsoft.com"
	const endpointPath = "/v1.0/users"

	// Mock logger
	mockLog := mocklogger.NewMockLogger()
	mockLog.On("Debug", mock.AnythingOfType("string"), mock.Anything).Once()

	handler := GraphAPIHandler{
		TenantID: "dummy-tenant-id",
		Logger:   mockLog,
	}

	// Set base domain assuming it's being mocked or controlled internally
	expectedURL := fmt.Sprintf("%s%s", baseURL, endpointPath)
	resultURL := handler.ConstructAPIResourceEndpoint(endpointPath, mockLog)

	assert.Equal(t, expectedURL, resultURL, "URL should match expected format")
	mockLog.AssertExpectations(t)
}

// TestConstructAPIAuthEndpoint tests the ConstructAPIAuthEndpoint function.
func TestConstructAPIAuthEndpoint(t *testing.T) {
	const baseURL = "https://login.microsoftonline.com"
	const endpointPath = "/oauth2/v2.0/token"

	// Mock logger
	mockLog := mocklogger.NewMockLogger()
	mockLog.On("Debug", mock.AnythingOfType("string"), mock.Anything).Once()

	handler := GraphAPIHandler{
		TenantID: "dummy-tenant-id",
		Logger:   mockLog,
	}

	// Construct the full URL by combining the base URL, tenant ID, and endpoint path.
	expectedURL := fmt.Sprintf("%s/%s%s", baseURL, handler.TenantID, endpointPath)
	resultURL := handler.ConstructAPIAuthEndpoint(endpointPath, mockLog)

	assert.Equal(t, expectedURL, resultURL, "URL should match expected format")
	mockLog.AssertExpectations(t)
}
