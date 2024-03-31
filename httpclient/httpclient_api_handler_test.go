// httpclient_api_handler.go
package httpclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/apihandlers/jamfpro"
	"github.com/deploymenttheory/go-api-http-client/apihandlers/msgraph"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// NewMockAPIHandler creates a new instance of MockAPIHandler.
func NewMockAPIHandler() *MockAPIHandler {
	return &MockAPIHandler{}
}

// Implement each method of the APIHandler interface on MockAPIHandler.

func (m *MockAPIHandler) ConstructAPIResourceEndpoint(instanceName string, endpointPath string, log logger.Logger) string {
	args := m.Called(instanceName, endpointPath, log)
	return args.String(0)
}

func (m *MockAPIHandler) ConstructAPIAuthEndpoint(instanceName string, endpointPath string, log logger.Logger) string {
	args := m.Called(instanceName, endpointPath, log)
	return args.String(0)
}

func (m *MockAPIHandler) MarshalRequest(body interface{}, method string, endpoint string, log logger.Logger) ([]byte, error) {
	args := m.Called(body, method, endpoint, log)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockAPIHandler) MarshalMultipartRequest(fields map[string]string, files map[string]string, log logger.Logger) ([]byte, string, error) {
	args := m.Called(fields, files, log)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *MockAPIHandler) HandleAPISuccessResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	args := m.Called(resp, out, log)
	return args.Error(0)
}

func (m *MockAPIHandler) HandleAPIErrorResponse(resp *http.Response, out interface{}, log logger.Logger) error {
	args := m.Called(resp, out, log)
	return args.Error(0)
}

func (m *MockAPIHandler) GetAPIBearerTokenAuthenticationSupportStatus() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAPIHandler) GetAPIOAuthAuthenticationSupportStatus() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAPIHandler) GetAPIOAuthWithCertAuthenticationSupportStatus() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAPIHandler) GetAcceptHeader() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAPIHandler) GetContentTypeHeader(method string, log logger.Logger) string {
	args := m.Called(method, log)
	return args.String(0)
}

func (m *MockAPIHandler) GetDefaultBaseDomain() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAPIHandler) GetOAuthTokenEndpoint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAPIHandler) GetBearerTokenEndpoint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAPIHandler) GetTokenRefreshEndpoint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAPIHandler) GetTokenInvalidateEndpoint() string {
	args := m.Called()
	return args.String(0)
}

// TestLoadAPIHandler verifies the functionality of the LoadAPIHandler function in the httpclient package.
// This function is designed to return the appropriate APIHandler implementation based on the provided apiType argument.
// The test cases cover the following scenarios:
// 1. Loading a JamfPro API handler by providing "jamfpro" as the apiType.
// 2. Loading a Graph API handler by providing "graph" as the apiType.
// 3. Handling an unsupported API type by providing an unknown apiType, which should result in an error.
func TestLoadAPIHandler(t *testing.T) {
	mockLogger := new(MockLogger)
	tests := []struct {
		name     string
		apiType  string
		wantType interface{}
		wantErr  bool
	}{
		{"Load JamfPro Handler", "jamfpro", &jamfpro.JamfAPIHandler{}, false},
		{"Load Graph Handler", "msgraph", &msgraph.GraphAPIHandler{}, false},
		{"Unsupported API Type", "unknown", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mockLogger.On("Error", mock.Anything, mock.Anything).Return(errors.New("Unsupported API type")).Once()
			} else {
				mockLogger.On("Info", mock.Anything, mock.Anything).Once()
			}

			got, err := LoadAPIHandler(tt.apiType, mockLogger)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.wantType, got)
			}

			mockLogger.AssertExpectations(t)
		})
	}
}
