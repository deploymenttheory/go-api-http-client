// http_helpers.go
package httpclient

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestParseISO8601Date tests the ParseISO8601Date function with various date strings
func TestParseISO8601Date(t *testing.T) {
	tests := []struct {
		dateStr      string
		expectErr    bool
		expectedTime time.Time // Add an expectedTime field for successful parsing
	}{
		{
			dateStr:      "2023-01-02T15:04:05Z",
			expectErr:    false,
			expectedTime: time.Date(2023, time.January, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			dateStr:      "2023-01-02T15:04:05-07:00",
			expectErr:    false,
			expectedTime: time.Date(2023, time.January, 2, 15, 4, 5, 0, time.FixedZone("", -7*3600)),
		},
		{
			dateStr:   "invalid-date",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.dateStr, func(t *testing.T) {
			result, err := ParseISO8601Date(tt.dateStr)

			if tt.expectErr {
				assert.Error(t, err, "Expected an error for date string: "+tt.dateStr)
			} else {
				assert.NoError(t, err, "Did not expect an error for date string: "+tt.dateStr)
				assert.True(t, result.Equal(tt.expectedTime), "Parsed time should match expected time")
			}
		})
	}
}

// TestRedactSensitiveData tests the RedactSensitiveData function with various scenarios
func TestRedactSensitiveData(t *testing.T) {
	tests := []struct {
		name            string
		hideSensitive   bool
		key             string
		value           string
		expectedOutcome string
	}{
		{"RedactSensitiveKey", true, "AccessToken", "secret-token", "REDACTED"},
		{"RedactSensitiveKeyAuthorization", true, "Authorization", "Bearer secret-token", "REDACTED"},
		{"DoNotRedactNonSensitiveKey", true, "NonSensitiveKey", "non-sensitive-value", "non-sensitive-value"},
		{"DoNotRedactWhenDisabled", false, "AccessToken", "secret-token", "secret-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				clientConfig: ClientConfig{
					ClientOptions: ClientOptions{
						HideSensitiveData: tt.hideSensitive,
					},
				},
			}

			result := RedactSensitiveHeaderData(client, tt.key, tt.value)
			assert.Equal(t, tt.expectedOutcome, result, "Redaction outcome should match expected")
		})
	}
}
