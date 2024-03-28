package httpclient

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
