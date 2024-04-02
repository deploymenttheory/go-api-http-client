// apiintegrations/apihandler/apihandler_test.go
package apihandler

import (
	"testing"

	"github.com/deploymenttheory/go-api-http-client/apiintegrations/jamfpro"
	"github.com/deploymenttheory/go-api-http-client/apiintegrations/msgraph"
	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoadAPIHandler(t *testing.T) {
	// Create a mock logger for testing purposes.
	mockLog := mocklogger.NewMockLogger()

	// Define your test cases.
	tests := []struct {
		name     string
		apiType  string
		wantType interface{}
		wantErr  bool
	}{
		{
			name:     "Load JamfPro Handler",
			apiType:  "jamfpro",
			wantType: &jamfpro.JamfAPIHandler{},
			wantErr:  false,
		},
		{
			name:     "Load Graph Handler",
			apiType:  "msgraph",
			wantType: &msgraph.GraphAPIHandler{},
			wantErr:  false,
		},
		{
			name:    "Unsupported API Type",
			apiType: "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup expectations for the mock logger based on whether an error is expected.
			if tt.wantErr {
				mockLog.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Once()
			} else {
				mockLog.On("Info", mock.Anything, mock.Anything, mock.Anything).Return().Once()
			}

			// Attempt to load the API handler.
			got, err := LoadAPIHandler(tt.apiType, mockLog)

			// Assert error handling.
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.wantType, got, "Got %T, want %T", got, tt.wantType)
			}

			// Assert that the mock logger's expectations were met.
			mockLog.AssertExpectations(t)
		})
	}
}
