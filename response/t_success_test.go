// response/success_test.go
package response

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleDeleteRequest_Success(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK, // Simulate a successful DELETE request
		Request: &http.Request{
			Method: "DELETE",
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/test",
			},
		},
	}

	err := successfulDeleteRequest(resp, nil)

	assert.NoError(t, err, "handleDeleteRequest should not return an error for successful DELETE requests")
}

func TestHandleDeleteRequest_Failure(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusBadRequest, // Simulate a failed DELETE request
		Request: &http.Request{
			Method: "DELETE",
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/test",
			},
		},
	}

	err := successfulDeleteRequest(resp, nil)

	assert.Error(t, err, "handleDeleteRequest should return an error for failed DELETE requests")
}
