/* When both the cookie jar is enabled and specific cookies are set for an HTTP client in
your scenario, here’s what generally happens during the request processing:

Cookie Jar Initialization: If the cookie jar is enabled through your SetupCookieJar function,
an instance of http.cookiejar.Jar is created and associated with your HTTP client. This
cookie jar will automatically handle incoming and outgoing cookies for all requests made
using this client. It manages storing cookies and automatically sending them with subsequent
requests to the domains for which they're valid.

Setting Specific Cookies: The ApplyCustomCookies function checks for any user-defined specific
cookies (from the CustomCookies map). If found, these cookies are explicitly added to the
outgoing HTTP request headers via the SetSpecificCookies function.

Interaction between Cookie Jar and Specific Cookies:
Cookie Precedence: When a specific cookie (added via SetSpecificCookies) shares the same name
as a cookie already handled by the cookie jar for a given domain, the behavior depends on the
implementation of the HTTP client's cookie handling and the server's cookie management rules.
Generally, the explicitly set cookie in the HTTP request header (from SetSpecificCookies)
should override any similar cookie managed by the cookie jar for that single request.

Subsequent Requests: For subsequent requests, if the specific cookies are not added again
via ApplyCustomCookies, the cookies in the jar that were stored from previous responses
will take precedence again, unless overwritten by subsequent responses or explicit setting
again.

Practical Usage:

This setup allows flexibility:

Use Cookie Jar: For general session management where cookies are automatically managed across
requests.
Use Specific Cookies: For overriding or adding specific cookies for particular requests where
customized control is necessary (such as testing scenarios, special authentication cookies,
etc.).
Logging and Debugging: Your setup also includes logging functionalities which can be very
useful to debug and verify which cookies are being sent and managed. This is crucial for
maintaining visibility into how cookies are influencing the behavior of your HTTP client
interactions.*/

package httpclient

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"net/url"
	"strings"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"go.uber.org/zap"
)

// SetupCookieJar initializes the HTTP client with a cookie jar if enabled in the configuration.

func SetupCookieJar(client *http.Client, clientConfig ClientConfig, log logger.Logger) error {
	if clientConfig.ClientOptions.Cookies.EnableCookieJar {

		jar, err := cookiejar.New(nil) // nil options use default options
		if err != nil {
			log.Error("Failed to create cookie jar", zap.Error(err))
			return fmt.Errorf("setupCookieJar failed: %w", err) // Wrap and return the error
		}

		if clientConfig.ClientOptions.Cookies.CustomCookies != nil {
			var CookieList []*http.Cookie
			CookieList = make([]*http.Cookie, 0)
			for k, v := range clientConfig.ClientOptions.Cookies.CustomCookies {
				newCookie := &http.Cookie{
					Name:  k,
					Value: v,
				}
				CookieList = append(CookieList, newCookie)
			}

			cookieUrl, err := url.Parse(fmt.Sprintf("http://%s.jamfcloud.com"))
			if err != nil {
				return err
			}

			jar.SetCookies(cookieUrl, CookieList)
		}

		client.Jar = jar
	}
	return nil
}

// ApplyCustomCookies checks and applies custom cookies to the HTTP request if any are configured.
// It logs the names of the custom cookies being applied without exposing their values.
func ApplyCustomCookies(req *http.Request, cookies map[string]string, log logger.Logger) {
	if len(cookies) > 0 {
		cookieNames := make([]string, 0, len(cookies))
		for name := range cookies {
			cookieNames = append(cookieNames, name)
		}
		log.Debug("Applying custom cookies", zap.Strings("Cookies", cookieNames))
		SetSpecificCookies(req, cookies)
	}
}

// SetSpecificCookies sets specific cookies provided in the configuration on the HTTP request.
func SetSpecificCookies(req *http.Request, cookies map[string]string) {
	for name, value := range cookies {
		cookie := &http.Cookie{
			Name:  name,
			Value: value,
		}
		req.AddCookie(cookie)
	}
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
