// cookiejar/cookiejar_test.go
package cookiejar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedactSensitiveCookies tests the RedactSensitiveCookies function to ensure it correctly redacts sensitive cookies.
func TestRedactSensitiveCookies(t *testing.T) {
	cookies := []*http.Cookie{
		{Name: "SessionID", Value: "sensitive-value-1"},
		{Name: "NonSensitiveCookie", Value: "non-sensitive-value"},
		{Name: "AnotherSensitiveCookie", Value: "sensitive-value-2"},
	}

	redactedCookies := RedactSensitiveCookies(cookies)

	// Define expected outcomes for each cookie.
	expectedValues := map[string]string{
		"SessionID":              "REDACTED",
		"NonSensitiveCookie":     "non-sensitive-value",
		"AnotherSensitiveCookie": "sensitive-value-2", // Assuming this is not in the sensitive list.
	}

	for _, cookie := range redactedCookies {
		assert.Equal(t, expectedValues[cookie.Name], cookie.Value, "Cookie value should match expected redaction outcome")
	}
}

// TestCookiesFromHeader tests the CookiesFromHeader function to ensure it can correctly parse cookies from HTTP headers.
func TestCookiesFromHeader(t *testing.T) {
	header := http.Header{
		"Set-Cookie": []string{
			"SessionID=sensitive-value; Path=/; HttpOnly",
			"NonSensitiveCookie=non-sensitive-value; Path=/",
		},
	}

	cookies := CookiesFromHeader(header)

	// Define expected outcomes for each cookie.
	expectedCookies := []*http.Cookie{
		{Name: "SessionID", Value: "sensitive-value"},
		{Name: "NonSensitiveCookie", Value: "non-sensitive-value"},
	}

	assert.Equal(t, len(expectedCookies), len(cookies), "Number of parsed cookies should match expected")

	for i, expectedCookie := range expectedCookies {
		assert.Equal(t, expectedCookie.Name, cookies[i].Name, "Cookie names should match")
		assert.Equal(t, expectedCookie.Value, cookies[i].Value, "Cookie values should match")
	}
}
