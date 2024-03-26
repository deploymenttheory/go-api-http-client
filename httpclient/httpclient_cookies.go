// httpclient_cookies.go
package httpclient

import (
	"net/http"
	"strings"
)

// RedactSensitiveCookies redacts sensitive information from cookies.
// It takes a slice of *http.Cookie and returns a redacted slice of *http.Cookie.
func RedactSensitiveCookies(cookies []*http.Cookie) []*http.Cookie {
	// Define sensitive cookie names that should be redacted.
	sensitiveCookieNames := map[string]bool{
		"SessionID": true, // Example sensitive cookie name
		// More sensitive cookie names will be added as needed.
	}

	// Iterate over the cookies and redact sensitive ones.
	for _, cookie := range cookies {
		if _, found := sensitiveCookieNames[cookie.Name]; found {
			cookie.Value = "REDACTED"
		}
	}

	return cookies
}

// Utility function to convert cookies from http.Header to []*http.Cookie.
// This can be useful if cookies are stored in http.Header (e.g., from a response).
func CookiesFromHeader(header http.Header) []*http.Cookie {
	cookies := []*http.Cookie{}
	for _, cookieHeader := range header["Set-Cookie"] {
		if cookie := ParseCookieHeader(cookieHeader); cookie != nil {
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

// ParseCookieHeader parses a single Set-Cookie header and returns an *http.Cookie.
func ParseCookieHeader(header string) *http.Cookie {
	headerParts := strings.Split(header, ";")
	if len(headerParts) > 0 {
		cookieParts := strings.SplitN(headerParts[0], "=", 2)
		if len(cookieParts) == 2 {
			return &http.Cookie{Name: cookieParts[0], Value: cookieParts[1]}
		}
	}
	return nil
}
