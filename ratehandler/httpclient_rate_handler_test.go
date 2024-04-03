// ratehandler/ratehandler.go
package ratehandler

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/deploymenttheory/go-api-http-client/mocklogger"
	"github.com/stretchr/testify/assert"
)

// TestCalculateBackoff tests the backoff calculation for various retry counts
func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		retry       int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{retry: 0, expectedMin: time.Duration(float64(baseDelay) * (1 - jitterFactor)), expectedMax: maxDelay},
		{retry: 1, expectedMin: time.Duration(float64(baseDelay*2) * (1 - jitterFactor)), expectedMax: maxDelay},
		{retry: 2, expectedMin: time.Duration(float64(baseDelay*4) * (1 - jitterFactor)), expectedMax: maxDelay},
		{retry: 5, expectedMin: time.Duration(float64(baseDelay*32) * (1 - jitterFactor)), expectedMax: maxDelay},
	}

	for _, tt := range tests {
		t.Run("RetryCount"+strconv.Itoa(tt.retry), func(t *testing.T) {
			delay := CalculateBackoff(tt.retry)

			// The delay should be within the expected range
			assert.GreaterOrEqual(t, delay, tt.expectedMin, "Delay should be greater than or equal to expected minimum after jitter adjustment")
			assert.LessOrEqual(t, delay, tt.expectedMax, "Delay should be less than or equal to expected maximum")
		})
	}
}

// TestParseRateLimitHeaders evaluates the functionality of the parseRateLimitHeaders function,
// ensuring it correctly interprets various rate-limiting headers from an HTTP response and calculates
// the appropriate wait duration. The function tests different scenarios including 'Retry-After' headers
// with both date and delay values, 'X-RateLimit-Reset' headers indicating the reset time for rate limiting,
// and situations where no relevant headers are present. Each test case mimics an HTTP response with specific
// headers set, and asserts that the calculated wait duration falls within an acceptable range of the expected
// value, allowing for slight variances due to execution time and rounding. The use of a mock logger ensures
// that the function's logging behavior can also be verified without affecting the output of the test runner.
func TestParseRateLimitHeaders(t *testing.T) {
	tests := []struct {
		name         string
		headers      map[string]string
		expectedWait time.Duration
	}{
		{
			name: "RetryAfterInSeconds",
			headers: map[string]string{
				"Retry-After": "120", // 2 minutes in seconds
			},
			expectedWait: 2 * time.Minute,
		},
		{
			name: "RetryAfterHTTPDate",
			headers: map[string]string{
				"Retry-After": time.Now().UTC().Format(time.RFC1123), // Use current time in RFC1123 format
			},
			expectedWait: 0, // Immediate retry since the date is current
		},
		{
			name: "XRateLimitReset",
			headers: map[string]string{
				"X-RateLimit-Remaining": "0",
				"X-RateLimit-Reset":     strconv.FormatInt(time.Now().Add(90*time.Second).Unix(), 10), // 90 seconds from now
			},
			expectedWait: 90*time.Second + 5*time.Second, // Add 5 seconds for skew buffer
		},
		{
			name:         "NoHeaders",
			headers:      map[string]string{},
			expectedWait: 0, // No wait since no headers are present
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{Header: http.Header{}}
			for k, v := range tt.headers {
				resp.Header.Add(k, v)
			}

			mockLog := mocklogger.NewMockLogger()
			wait := ParseRateLimitHeaders(resp, mockLog)

			// Adjust the delta based on the expected wait duration
			delta := time.Duration(1) * time.Second
			if tt.expectedWait == 0 {
				// For immediate retries, allow a larger delta
				delta = time.Duration(5) * time.Second
			}

			assert.InDelta(t, tt.expectedWait, wait, float64(delta), "Wait duration should be within expected range")
		})
	}
}
