// httpclient_error_response_test.go
// This package provides utility functions and structures for handling and categorizing HTTP error responses.
package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHandleAPIErrorResponse tests the handleAPIErrorResponse function with different types of error responses.
func TestHandleAPIErrorResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       *http.Response
		expectedAPIErr *APIError
	}{
		{
			name: "Structured JSON Error",
			response: &http.Response{
				StatusCode: 400,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":{"code":"INVALID","message":"Invalid request"}}`)),
			},
			expectedAPIErr: &APIError{
				StatusCode: 400,
				Type:       "APIError",
				Message:    "Invalid request",
				Detail:     "",
				Errors:     nil,
				Raw:        `{"error":{"code":"INVALID","message":"Invalid request"}}`,
			},
		},
		{
			name: "Generic JSON Error",
			response: &http.Response{
				StatusCode: 500,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Internal server error","detail":"Error details"}`)),
			},
			expectedAPIErr: &APIError{
				StatusCode: 500,
				Type:       "APIError",
				Message:    "Internal server error",
				Detail:     "Error details",
				Errors:     nil,
				Raw:        `{"message":"Internal server error","detail":"Error details"}`,
			},
		},
		{
			name: "Non-JSON Error",
			response: &http.Response{
				StatusCode: 404,
				Header:     http.Header{"Content-Type": []string{"text/html"}},
				Body:       io.NopCloser(bytes.NewBufferString("<html>Not Found</html>")),
			},
			expectedAPIErr: &APIError{
				StatusCode: 404,
				Type:       "APIError",
				Message:    "An error occurred",
				Detail:     "",
				Errors:     nil,
				Raw:        "<html>Not Found</html>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := NewMockLogger()
			apiError := handleAPIErrorResponse(tt.response, mockLog)

			assert.Equal(t, tt.expectedAPIErr.StatusCode, apiError.StatusCode)
			assert.Equal(t, tt.expectedAPIErr.Message, apiError.Message)
			assert.Equal(t, tt.expectedAPIErr.Detail, apiError.Detail)
			assert.Equal(t, tt.expectedAPIErr.Raw, apiError.Raw)
		})
	}
}
