// cookiejar/cookiejar.go

/* The cookiejar package provides utility functions for managing cookies within an HTTP client
context in Go. This package aims to enhance HTTP client functionalities by offering cookie
handling capabilities, including initialization of a cookie jar, redaction of sensitive cookies,
and parsing of cookies from HTTP headers. Below is an explanation of the core functionalities
provided by this package*/

package cookiejar

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// SetupCookieJar initializes the HTTP client with a cookie jar if enabled in the configuration.
func SetupCookieJar(client *http.Client, enableCookieJar bool, log logger.Logger) error {
	if enableCookieJar {
		jar, err := cookiejar.New(nil) // nil options use default options
		if err != nil {
			log.Error("Failed to create cookie jar", zap.Error(err))
			return fmt.Errorf("setupCookieJar failed: %w", err) // Wrap and return the error
		}
		client.Jar = jar
	}
	return nil
}

// GetCookies is a middleware that extracts cookies from incoming requests and serializes them.
func GetCookies(next http.Handler, log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Extract cookies from the request
		cookies := r.Cookies()

		// Serialize the cookies
		serializedCookies := SerializeCookies(cookies)

		// Log the serialized cookies
		log.Info("Serialized Cookies", zap.String("Cookies", serializedCookies))

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// SerializeCookies serializes a slice of *http.Cookie into a string format.
func SerializeCookies(cookies []*http.Cookie) string {
	var cookieStrings []string

	for _, cookie := range cookies {
		cookieStrings = append(cookieStrings, cookie.String())
	}

	return strings.Join(cookieStrings, "; ")
}

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
