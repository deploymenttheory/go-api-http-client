// headers/redact/redact_test.go
package redact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedactSensitiveHeaderData tests the RedactSensitiveHeaderData function to ensure it correctly redacts sensitive data.
func TestRedactSensitiveHeaderData(t *testing.T) {
	cases := []struct {
		name              string
		hideSensitiveData bool
		key               string
		value             string
		expected          string
	}{
		{"Sensitive Key With Redaction", true, "AccessToken", "some-sensitive-token", "REDACTED"},
		{"Sensitive Key Without Redaction", false, "AccessToken", "some-sensitive-token", "some-sensitive-token"},
		{"Non-Sensitive Key With Redaction", true, "User-Agent", "MyCustomAgent", "MyCustomAgent"},
		{"Non-Sensitive Key Without Redaction", false, "User-Agent", "MyCustomAgent", "MyCustomAgent"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := RedactSensitiveHeaderData(tc.hideSensitiveData, tc.key, tc.value)
			assert.Equal(t, tc.expected, result, "Redacted value should match the expected outcome")
		})
	}
}
