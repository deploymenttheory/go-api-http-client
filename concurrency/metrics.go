package concurrency

import (
	"math"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// MonitorRateLimitHeaders monitors the rate limit headers (X-RateLimit-Remaining and Retry-After)
// in the HTTP response and adjusts concurrency accordingly.
func (ch *ConcurrencyHandler) MonitorRateLimitHeaders(resp *http.Response) {
	// Extract X-RateLimit-Remaining and Retry-After headers from the response
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	retryAfter := resp.Header.Get("Retry-After")

	// Adjust concurrency based on the values of these headers
	// Implement your logic here to dynamically adjust concurrency
}

// MonitorServerResponseCodes monitors server response codes and adjusts concurrency accordingly.
func (ch *ConcurrencyHandler) MonitorServerResponseCodes(resp *http.Response) {
	statusCode := resp.StatusCode
	// Check for 5xx errors (server errors) and 4xx errors (client errors)
	// Implement your logic here to track increases in error rates and adjust concurrency
}

// MonitorResponseTimeVariability calculates the standard deviation of response times
// and uses moving averages to smooth out fluctuations, adjusting concurrency accordingly.
func (ch *ConcurrencyHandler) MonitorResponseTimeVariability(responseTime time.Duration) {
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Update TotalResponseTime and ResponseCount for moving average calculation
	ch.Metrics.TotalResponseTime += responseTime
	ch.Metrics.ResponseCount++

	// Calculate average response time
	averageResponseTime := ch.Metrics.TotalResponseTime / time.Duration(ch.Metrics.ResponseCount)

	// Calculate standard deviation of response times
	variance := ch.calculateVariance(averageResponseTime, responseTime)
	stdDev := math.Sqrt(variance)

	// Adjust concurrency based on response time variability
	if float64(stdDev) > MaxAcceptableResponseTimeVariability.Seconds() && len(ch.sem) > MinConcurrency {
		newSize := len(ch.sem) - 1
		ch.logger.Info("Reducing concurrency due to high response time variability", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	}
}

// calculateVariance calculates the variance of response times.
func (ch *ConcurrencyHandler) calculateVariance(averageResponseTime time.Duration, responseTime time.Duration) float64 {
	// Convert time.Duration values to seconds
	averageSeconds := averageResponseTime.Seconds()
	responseSeconds := responseTime.Seconds()

	// Calculate variance
	variance := (float64(ch.Metrics.ResponseCount-1)*math.Pow(averageSeconds-responseSeconds, 2) + ch.Metrics.Variance) / float64(ch.Metrics.ResponseCount)
	ch.Metrics.Variance = variance
	return variance
}

// MonitorNetworkLatency measures Time to First Byte (TTFB) and monitors network throughput,
// adjusting concurrency based on changes in network latency and throughput.
func (ch *ConcurrencyHandler) MonitorNetworkLatency(ttfb time.Duration, throughput float64) {
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Calculate the TTFB moving average
	ch.Metrics.TTFB.Lock.Lock()
	defer ch.Metrics.TTFB.Lock.Unlock()
	ch.Metrics.TTFB.Total += ttfb
	ch.Metrics.TTFB.Count++
	ttfbMovingAverage := ch.Metrics.TTFB.Total / time.Duration(ch.Metrics.TTFB.Count)

	// Calculate the throughput moving average
	ch.Metrics.Throughput.Lock.Lock()
	defer ch.Metrics.Throughput.Lock.Unlock()
	ch.Metrics.Throughput.Total += throughput
	ch.Metrics.Throughput.Count++
	throughputMovingAverage := ch.Metrics.Throughput.Total / float64(ch.Metrics.Throughput.Count)

	// Adjust concurrency based on TTFB and throughput moving averages
	if ttfbMovingAverage > MaxAcceptableTTFB && len(ch.sem) > MinConcurrency {
		newSize := len(ch.sem) - 1
		ch.logger.Info("Reducing concurrency due to high TTFB", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else if throughputMovingAverage > MaxAcceptableThroughput && len(ch.sem) < MaxConcurrency {
		newSize := len(ch.sem) + 1
		ch.logger.Info("Increasing concurrency due to high throughput", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	}
}
