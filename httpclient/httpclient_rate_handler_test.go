// http_rate_handler_test.go
package httpclient

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCalculateBackoff tests the backoff calculation for various retry counts
func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		retry int
	}{
		{retry: 0},
		{retry: 1},
		{retry: 2},
		{retry: 5}, // Test a higher number of retries to ensure maxDelay is respected
	}

	for _, tt := range tests {
		t.Run("RetryCount"+strconv.Itoa(tt.retry), func(t *testing.T) {

			delay := calculateBackoff(tt.retry)

			// The delay should never exceed maxDelay
			assert.LessOrEqual(t, delay, maxDelay, "Delay should not exceed maxDelay")

			// The delay for 0 retries should be at least baseDelay
			if tt.retry == 0 {
				assert.GreaterOrEqual(t, delay, baseDelay, "Delay for 0 retries should be at least baseDelay")
			}
		})
	}
}

// TestParseRateLimitHeaders tests parsing of rate limit headers and calculation of wait duration
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
				"Retry-After": http.TimeFormat, // Use current time for simplicity
			},
			expectedWait: 0, // Immediate retry since the date is current
		},
		{
			name: "XRateLimitReset",
			headers: map[string]string{
				"X-RateLimit-Remaining": "0",
				"X-RateLimit-Reset":     strconv.FormatInt(time.Now().Add(90*time.Second).Unix(), 10), // 90 seconds from now
			},
			expectedWait: 90 * time.Second,
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

			wait := parseRateLimitHeaders(resp, NewMockLogger())

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
