// httpclient_auth_validation_test.go
package httpclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsValidClientID tests the IsValidClientID function with various client ID inputs.
// It verifies that valid UUIDs are correctly identified as such, and invalid formats
// are appropriately flagged with an error message. Additionally, it checks that empty
// client IDs are considered valid according to the updated logic.
func TestIsValidClientID(t *testing.T) {
	tests := []struct {
		clientID    string
		expected    bool
		expectedMsg string
	}{
		{"123e4567-e89b-12d3-a456-426614174000", true, ""},
		{"invalid-uuid", false, "Client ID is not a valid UUID format."},
		{"", true, ""}, // Empty client ID should be considered valid as per your updated logic
	}

	for _, tt := range tests {
		valid, msg := IsValidClientID(tt.clientID)
		assert.Equal(t, tt.expected, valid)
		assert.Equal(t, tt.expectedMsg, msg)
	}
}

// TestIsValidClientSecret tests the IsValidClientSecret function with various client secret inputs.
// It ensures that client secrets that meet the minimum length requirement and contain the necessary
// character types are validated correctly. It also checks that short or invalid client secrets are
// flagged appropriately, and that empty client secrets are considered valid as per the updated logic.
func TestIsValidClientSecret(t *testing.T) {
	tests := []struct {
		clientSecret string
		expected     bool
		expectedMsg  string
	}{
		{"ValidSecret123!", true, ""},
		{"short", false, "Client secret must be at least 16 characters long."},
		{"", true, ""}, // Empty client secret should be considered valid as per your updated logic
	}

	for _, tt := range tests {
		valid, msg := IsValidClientSecret(tt.clientSecret)
		assert.Equal(t, tt.expected, valid)
		assert.Equal(t, tt.expectedMsg, msg)
	}
}

// TestIsValidUsername tests the IsValidUsername function with various username inputs.
// This function verifies that usernames consisting of alphanumeric characters and password safe
// special characters are considered valid. It also checks that usernames with unsafe characters
// are correctly identified as invalid.
func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		username    string
		expected    bool
		expectedMsg string
	}{
		{"user123", true, ""},
		{"user!@#", true, ""},
		{"<script>", false, "Username must contain only alphanumeric characters and password safe special characters (!@#$%^&*()_-+=[{]}\\|;:'\",<.>/?)."},
	}

	for _, tt := range tests {
		valid, msg := IsValidUsername(tt.username)
		assert.Equal(t, tt.expected, valid)
		assert.Equal(t, tt.expectedMsg, msg)
	}
}

// TestIsValidPassword tests the IsValidPassword function with various password inputs.
// It ensures that passwords meeting the minimum length requirement are validated correctly,
// and that short passwords are appropriately flagged as invalid.
func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		password    string
		expected    bool
		expectedMsg string
	}{
		{"Password1", true, ""},
		{"short", false, "Password must be at least 8 characters long."},
	}

	for _, tt := range tests {
		valid, msg := IsValidPassword(tt.password)
		assert.Equal(t, tt.expected, valid)
		assert.Equal(t, tt.expectedMsg, msg)
	}
}

// TestDetermineAuthMethod tests the DetermineAuthMethod function with various authentication configurations.
// It checks that the function correctly identifies the authentication method to be used based on the provided
// credentials. Scenarios include valid OAuth credentials, valid bearer token credentials, and various combinations
// of invalid or missing credentials. The function should return "oauth" for valid OAuth credentials, "bearer" for
// valid bearer token credentials, and "unknown" with an error message for invalid or incomplete credentials.
func TestDetermineAuthMethod(t *testing.T) {
	tests := []struct {
		authConfig  AuthConfig
		expected    string
		expectedErr error
	}{
		{AuthConfig{ClientID: "123e4567-e89b-12d3-a456-426614174000", ClientSecret: "ValidSecret123!"}, "oauth", nil},
		{AuthConfig{Username: "user123", Password: "Password1"}, "bearer", nil},
		{AuthConfig{ClientID: "invalid-uuid", ClientSecret: "ValidSecret123!"}, "unknown", errors.New("No valid credentials provided. Client ID is not a valid UUID format.")},
		{AuthConfig{}, "unknown", errors.New("No valid credentials provided.")}, // No credentials provided
	}

	for _, tt := range tests {
		method, err := DetermineAuthMethod(tt.authConfig)
		assert.Equal(t, tt.expected, method)
		if tt.expectedErr != nil {
			assert.EqualError(t, err, tt.expectedErr.Error())
		} else {
			assert.NoError(t, err)
		}
	}
}
