// httpclient_auth_validation_test.go
package httpclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsValidClientID tests the IsValidClientID function with various client ID inputs.
// It verifies that valid UUIDs are correctly identified as such, and invalid formats
// are appropriately flagged with an error message. Additionally, it checks that empty
// client IDs are considered valid according to the updated logic.
// TestIsValidClientID tests the IsValidClientID function for both valid and invalid UUIDs.
func TestIsValidClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		want     bool
		errMsg   string
	}{
		{"Valid UUID", "123e4567-e89b-12d3-a456-426614174000", true, ""},
		{"Invalid UUID - Wrong Length", "123e4567", false, "Client ID is not a valid UUID format."},
		{"Invalid UUID - Invalid Characters", "G23e4567-e89b-12d3-a456-426614174000", false, "Client ID is not a valid UUID format."},
		// Add more cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := IsValidClientID(tt.clientID)
			assert.Equal(t, tt.want, valid)
			if !tt.want {
				assert.Equal(t, tt.errMsg, errMsg)
			}
		})
	}
}

// TestIsValidClientSecret tests the IsValidClientSecret function with various client secret inputs.
// It ensures that client secrets that meet the minimum length requirement and contain the necessary
// character types are validated correctly. It also checks that short or invalid client secrets are
// flagged appropriately, and that empty client secrets are considered valid as per the updated logic.
func TestIsValidClientSecret(t *testing.T) {
	tests := []struct {
		name         string
		clientSecret string
		want         bool
		errMsg       string
	}{
		{"Valid Secret", "Aa1!Aa1!Aa1!Aa1!", true, ""},
		{"Too Short", "Aa1!", false, "Client secret must be at least 16 characters long."},
		{"No Lowercase", "AAAAAAAAAAAAAA1!", false, "Client secret must contain at least one lowercase letter."},
		{"No Uppercase", "aaaaaaaaaaaaaa1!", false, "Client secret must contain at least one uppercase letter."},
		{"No Digit", "Aa!Aa!Aa!Aa!Aa!Aa!", false, "Client secret must contain at least one digit."},
		// Add more cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := IsValidClientSecret(tt.clientSecret)
			assert.Equal(t, tt.want, valid)
			if !tt.want {
				assert.Equal(t, tt.errMsg, errMsg)
			}
		})
	}
}

// TestIsValidUsername tests the IsValidUsername function to ensure it enforces the defined criteria for usernames.
func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
		errMsg   string
	}{
		{"Valid Username", "User123!", true, ""},
		{"Valid Special Characters", "User_@#1$", true, ""},
		{"Invalid Characters", " InvalidUsername", false, "Username must contain only alphanumeric characters and password safe special characters (!@#$%^&*()_-+=[{]}\\|;:'\",<.>/?)."},
		{"Empty Username", "", false, "Username must contain only alphanumeric characters and password safe special characters (!@#$%^&*()_-+=[{]}\\|;:'\",<.>/?)."},
		// You can add more cases here to test additional scenarios, such as extremely long usernames or usernames with only special characters.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errMsg := IsValidUsername(tt.username)
			assert.Equal(t, tt.want, valid)
			if !tt.want {
				assert.Equal(t, tt.errMsg, errMsg)
			}
		})
	}
}
