package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock type for the Logger interface, useful for testing without needing a real logger implementation.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...interface{}) error {
	args := m.Called(msg, fields)
	return args.Error(0)
}

// TestHandleAPIErrorResponse tests the handling of various API error responses.
func TestHandleAPIErrorResponse(t *testing.T) {
	tests := []struct {
		name             string
		responseStatus   int
		responseBody     string
		expectedAPIError *APIError
	}{
		{
			name:           "structured JSON error",
			responseStatus: http.StatusBadRequest,
			responseBody:   `{"error": {"code": "400", "message": "Bad Request"}}`,
			expectedAPIError: &APIError{
				StatusCode: http.StatusBadRequest,
				Type:       "APIError",
				Message:    "An error occurred",
				Raw:        `{"error": {"code": "400", "message": "Bad Request"}}`,
			},
		},
		{
			name:           "generic JSON error",
			responseStatus: http.StatusInternalServerError,
			responseBody:   `{"message": "Internal Server Error", "detail": "Server crashed"}`,
			expectedAPIError: &APIError{
				StatusCode: http.StatusInternalServerError,
				Type:       "APIError",
				Message:    "An error occurred",
				Raw:        `{"message": "Internal Server Error", "detail": "Server crashed"}`,
			},
		},
		{
			name:           "non-JSON error",
			responseStatus: http.StatusServiceUnavailable,
			responseBody:   `<html><body>Service Unavailable</body></html>`,
			expectedAPIError: &APIError{
				StatusCode: http.StatusServiceUnavailable,
				Type:       "APIError",
				Message:    "An error occurred",
				Raw:        `<html><body>Service Unavailable</body></html>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP response
			responseRecorder := httptest.NewRecorder()
			responseRecorder.WriteHeader(tt.responseStatus)
			responseRecorder.WriteString(tt.responseBody)

			// Create a dummy request and associate it with the response
			dummyReq := httptest.NewRequest("GET", "http://example.com", nil)
			response := responseRecorder.Result()
			response.Request = dummyReq

			// Use the centralized MockLogger from the mocklogger package
			mockLogger := mocklogger.NewMockLogger()

			// Set up expectations for LogError
			mockLogger.On("LogError",
				mock.AnythingOfType("string"), // event
				mock.AnythingOfType("string"), // method
				mock.AnythingOfType("string"), // url
				mock.AnythingOfType("int"),    // statusCode
				mock.AnythingOfType("string"), // status
				mock.Anything,                 // error
				mock.AnythingOfType("string"), // raw response
			).Return()

			// Call HandleAPIErrorResponse
			result := HandleAPIErrorResponse(response, mockLogger)

			// Assert
			assert.Equal(t, tt.expectedAPIError.StatusCode, result.StatusCode)
			assert.Equal(t, tt.expectedAPIError.Type, result.Type)
			assert.Equal(t, tt.expectedAPIError.Raw, result.Raw)

			// Assert that all expectations were met
			mockLogger.AssertExpectations(t)
		})
	}
}
