// http_request_test.go
package httpclient

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockClient extends httpclient.Client to override methods for testing
type MockClient struct {
	Client
	Mock mock.Mock
}

// executeRequestWithRetries mock
func (m *MockClient) executeRequestWithRetries(method, endpoint string, body, out interface{}, log logger.Logger) (*http.Response, error) {
	args := m.Mock.Called(method, endpoint, body, out, log)
	return args.Get(0).(*http.Response), args.Error(1)
}

// executeRequest mock
func (m *MockClient) executeRequest(method, endpoint string, body, out interface{}, log logger.Logger) (*http.Response, error) {
	args := m.Mock.Called(method, endpoint, body, out, log)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestDoRequest(t *testing.T) {
	mockLogger := NewMockLogger()
	testClient := &MockClient{}

	// Mock responses
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
	}

	// Define test cases
	testCases := []struct {
		method        string
		useRetries    bool
		expectedError error
	}{
		{"GET", true, nil},
		{"PUT", true, nil},
		{"DELETE", true, nil},
		{"POST", false, nil},
		{"PATCH", false, nil},
		{"UNSUPPORTED", false, mockLogger.Error("HTTP method not supported", zap.String("method", "UNSUPPORTED"))},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			if tc.useRetries {
				testClient.Mock.On("executeRequestWithRetries", tc.method, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, tc.expectedError)
			} else {
				testClient.Mock.On("executeRequest", tc.method, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, tc.expectedError)
			}

			resp, err := testClient.DoRequest(tc.method, "/test", nil, nil, mockLogger)

			if tc.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}

			if tc.useRetries {
				testClient.Mock.AssertCalled(t, "executeRequestWithRetries", tc.method, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			} else if tc.method != "UNSUPPORTED" {
				testClient.Mock.AssertCalled(t, "executeRequest", tc.method, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestExecuteRequestWithRetries(t *testing.T) {
	mockLogger := NewMockLogger()
	testClient := &MockClient{}

	// Simulate transient error response
	transientErrorResponse := &http.Response{StatusCode: http.StatusServiceUnavailable}

	testCases := []struct {
		name            string
		response        *http.Response
		expectedRetries int
		expectedError   error
	}{
		{
			name:            "TransientErrorWithRetry",
			response:        transientErrorResponse,
			expectedRetries: 3, // Assuming max retry attempts is 3
			expectedError:   nil,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testClient.Mock.On("executeHTTPRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.response, errors.New("transient error")).Times(tc.expectedRetries)
			testClient.Mock.On("executeHTTPRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&http.Response{StatusCode: http.StatusOK}, nil).Once()

			_, err := testClient.executeRequestWithRetries("GET", "/test", nil, nil, mockLogger)

			if tc.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			testClient.Mock.AssertNumberOfCalls(t, "executeHTTPRequest", tc.expectedRetries+1)
		})
	}
}

// mockHTTPClient is a mock of the http.Client
type mockHTTPClient struct {
	DoFunc func(*http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestHandleErrorResponse(t *testing.T) {
	mockLogger := NewMockLogger()
	client := &Client{
		APIHandler: NewMockAPIHandler(), // Assume NewMockAPIHandler returns a mock that satisfies your APIHandler interface
	}

	// Simulate an API error response
	apiErrorResponse := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`{"error":"bad request"}`)),
	}

	err := client.handleErrorResponse(apiErrorResponse, nil, mockLogger, "POST", "/test")
	assert.Error(t, err)
	// Additional assertions based on the error handling logic in your APIHandler
}
