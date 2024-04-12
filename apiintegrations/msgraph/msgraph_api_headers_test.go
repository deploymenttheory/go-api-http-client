// apiintegrations/msgraph/msgraph_api_headers_test.go
package msgraph

import (
	"testing"

	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGetAPIRequestHeaders tests the GetAPIRequestHeaders function.
func TestGetAPIRequestHeaders(t *testing.T) {
	handler := GraphAPIHandler{Logger: mocklogger.NewMockLogger()}
	endpoint := "/api/data"
	handler.Logger.(*mocklogger.MockLogger).On("Debug", mock.Anything, mock.Anything).Maybe()

	expectedHeaders := map[string]string{
		"Accept":        "application/x-x509-ca-cert;q=0.95,application/pkix-cert;q=0.94,application/pem-certificate-chain;q=0.93,application/octet-stream;q=0.8,image/png;q=0.75,image/jpeg;q=0.74,image/*;q=0.7,application/xml;q=0.65,text/xml;q=0.64,text/xml;charset=UTF-8;q=0.63,application/json;q=0.5,text/html;q=0.5,text/plain;q=0.4,*/*;q=0.05",
		"Content-Type":  "application/json",
		"Authorization": "",
		"User-Agent":    "go-api-http-client-msgraph-handler",
	}

	headers := handler.GetAPIRequestHeaders(endpoint)
	assert.Equal(t, expectedHeaders, headers)
	handler.Logger.(*mocklogger.MockLogger).AssertExpectations(t)
}

// TestGetContentTypeHeader tests the GetContentTypeHeader function.
func TestGetAcceptHeader(t *testing.T) {
	handler := GraphAPIHandler{}
	acceptHeader := handler.GetAcceptHeader()
	expectedHeader := "application/x-x509-ca-cert;q=0.95,application/pkix-cert;q=0.94,application/pem-certificate-chain;q=0.93,application/octet-stream;q=0.8,image/png;q=0.75,image/jpeg;q=0.74,image/*;q=0.7,application/xml;q=0.65,text/xml;q=0.64,text/xml;charset=UTF-8;q=0.63,application/json;q=0.5,text/html;q=0.5,text/plain;q=0.4,*/*;q=0.05"
	assert.Equal(t, expectedHeader, acceptHeader, "The Accept header should correctly prioritize MIME types.")
}
