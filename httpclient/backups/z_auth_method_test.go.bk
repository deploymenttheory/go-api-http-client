package httpclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineAuthMethod(t *testing.T) {
	tests := []struct {
		name         string
		authConfig   AuthConfig
		expectedAuth string
		expectError  bool
	}{
		{
			name: "Valid OAuth credentials",
			authConfig: AuthConfig{
				ClientID:     "123e4567-e89b-12d3-a456-426614174000", // Valid UUID format
				ClientSecret: "validSecretWith16Chars",               // Ensure it's at least 16 characters
			},
			expectedAuth: "oauth",
			expectError:  false,
		},
		{
			name: "Valid Bearer credentials",
			authConfig: AuthConfig{
				Username: "validUsername",
				Password: "validPassword",
			},
			expectedAuth: "bearer",
			expectError:  false,
		},
		{
			name: "Invalid OAuth credentials",
			authConfig: AuthConfig{
				ClientID:     "invalidClientID",
				ClientSecret: "invalidClientSecret",
			},
			expectedAuth: "unknown",
			expectError:  true,
		},
		{
			name: "Invalid Bearer credentials",
			authConfig: AuthConfig{
				Username: "invalidUser",
				Password: "short",
			},
			expectedAuth: "unknown",
			expectError:  true,
		},
		{
			name:         "Missing credentials",
			authConfig:   AuthConfig{},
			expectedAuth: "unknown",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authMethod, err := DetermineAuthMethod(tt.authConfig)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAuth, authMethod)
		})
	}
}
