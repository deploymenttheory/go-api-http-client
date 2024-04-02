// headers/headers_test.go
package headers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deploymenttheory/go-api-http-client/authenticationhandler"
	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetAuthorization(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	mockLog := mocklogger.NewMockLogger()
	mockLog.On("Debug", mock.Anything, mock.Anything).Once()

	token := "test-token"
	authTokenHandler := &authenticationhandler.AuthTokenHandler{Token: token}

	headerHandler := NewHeaderHandler(req, mockLog, nil, authTokenHandler)
	headerHandler.SetAuthorization()

	assert.Equal(t, "Bearer "+token, req.Header.Get("Authorization"), "Authorization header should be correctly set")
	mockLog.AssertExpectations(t)
}

func TestSetContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	mockLog := mocklogger.NewMockLogger()

	contentType := "application/json"
	headerHandler := NewHeaderHandler(req, mockLog, nil, nil)
	headerHandler.SetContentType(contentType)

	assert.Equal(t, contentType, req.Header.Get("Content-Type"), "Content-Type header should be correctly set")
}

func TestLogHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	mockLog := mocklogger.NewMockLogger()
	mockLog.On("Debug", mock.Anything, mock.Anything).Once()
	mockLog.SetLevel(logger.LogLevelDebug)

	headerHandler := NewHeaderHandler(req, mockLog, nil, nil)
	headerHandler.LogHeaders(true)

	mockLog.AssertExpectations(t)
}
