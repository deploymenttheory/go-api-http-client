// ratehandler/ratehandler.go

/*
Components:

Backoff Strategy: A function that calculates the delay before the next retry. It will
implement exponential backoff with jitter. This strategy is more effective than a fixed
delay, as it ensures that in cases of prolonged issues, the client won't keep hammering
the server with a high frequency.
Response Time Monitoring: We'll introduce a mechanism to track average response times and
use deviations from this average to inform our backoff strategy.
Error Classifier: A function to classify different types of errors. Only transient errors
should be retried.Rate Limit Header Parser: For future compatibility, a function that can
parse common rate limit headers (like X-RateLimit-Remaining and Retry-After) and adjust
behavior accordingly.

*/

package ratehandler

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Constants for exponential backoff with jitter
const (
	baseDelay    = 100 * time.Millisecond // Initial delay
	maxDelay     = 5 * time.Second        // Maximum delay
	jitterFactor = 0.5                    // Random jitter factor
)

// CalculateBackoff calculates the next delay for retry with exponential backoff and jitter.
// The baseDelay is the initial delay duration, which is exponentially increased on each retry.
// The jitterFactor adds randomness to the delay to avoid simultaneous retries (thundering herd problem).
// The delay is capped at maxDelay to prevent excessive wait times.
func CalculateBackoff(retry int) time.Duration {
	delay := float64(baseDelay) * math.Pow(2, float64(retry))

	jitter := (rand.Float64() - 0.5) * jitterFactor * 2.0 // Random value between -jitterFactor and +jitterFactor
	delayWithJitter := delay * (1.0 + jitter)

	if delayWithJitter > float64(maxDelay) {
		return maxDelay
	}
	return time.Duration(delayWithJitter)
}

// ParseRateLimitHeaders parses common rate limit headers and adjusts behavior accordingly.
// It handles both Retry-After (in seconds or HTTP-date format) and X-RateLimit-Reset headers.
func ParseRateLimitHeaders(resp *http.Response, logger *zap.SugaredLogger) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		if waitSeconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(waitSeconds) * time.Second

		} else if retryAfterDate, err := time.Parse(time.RFC1123, retryAfter); err == nil {
			return time.Until(retryAfterDate)

		} else {
			logger.Debug("Unable to parse Retry-After header", zap.String("value", retryAfter), zap.Error(err))
		}
	}

	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
		if resetTimeStr := resp.Header.Get("X-RateLimit-Reset"); resetTimeStr != "" {
			if resetTimeEpoch, err := strconv.ParseInt(resetTimeStr, 10, 64); err == nil {
				resetTime := time.Unix(resetTimeEpoch, 0)
				return time.Until(resetTime) + (5 * time.Second)

			} else {
				logger.Debug("Unable to parse X-RateLimit-Reset header", zap.String("value", resetTimeStr), zap.Error(err))
			}
		}
	}

	return 0
}
