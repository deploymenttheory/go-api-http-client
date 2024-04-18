package concurrency

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// MonitorRateLimitHeaders monitors the rate limit headers (X-RateLimit-Remaining and Retry-After)
// in the HTTP response and adjusts concurrency accordingly.
// If X-RateLimit-Remaining is below a threshold or Retry-After is specified, decrease concurrency.
// If neither condition is met, consider scaling up if concurrency is below the maximum limit.
// - Threshold for X-RateLimit-Remaining: 10
// - Maximum concurrency: MaxConcurrency
func (ch *ConcurrencyHandler) MonitorRateLimitHeaders(resp *http.Response) {
	// Extract X-RateLimit-Remaining and Retry-After headers from the response
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	retryAfter := resp.Header.Get("Retry-After")

	if remaining != "" {
		remainingValue, err := strconv.Atoi(remaining)
		if err == nil && remainingValue < 10 {
			// Decrease concurrency if X-RateLimit-Remaining is below the threshold
			if len(ch.sem) > MinConcurrency {
				newSize := len(ch.sem) - 1
				ch.logger.Info("Reducing concurrency due to low X-RateLimit-Remaining", zap.Int("NewSize", newSize))
				ch.ResizeSemaphore(newSize)
			}
		}
	}

	if retryAfter != "" {
		// Decrease concurrency if Retry-After is specified
		if len(ch.sem) > MinConcurrency {
			newSize := len(ch.sem) - 1
			ch.logger.Info("Reducing concurrency due to Retry-After header", zap.Int("NewSize", newSize))
			ch.ResizeSemaphore(newSize)
		}
	} else {
		// Scale up if concurrency is below the maximum limit
		if len(ch.sem) < MaxConcurrency {
			newSize := len(ch.sem) + 1
			ch.logger.Info("Increasing concurrency", zap.Int("NewSize", newSize))
			ch.ResizeSemaphore(newSize)
		}
	}
}

// MonitorServerResponseCodes monitors server response codes and adjusts concurrency accordingly.
func (ch *ConcurrencyHandler) MonitorServerResponseCodes(resp *http.Response) {
	statusCode := resp.StatusCode

	// Lock the metrics to ensure thread safety
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Update the appropriate error count based on the response status code
	switch {
	case statusCode >= 500 && statusCode < 600:
		ch.Metrics.TotalRateLimitErrors++
	case statusCode >= 400 && statusCode < 500:
		// Assuming 4xx errors as client errors
		// Increase the TotalRetries count to indicate a client error
		ch.Metrics.TotalRetries++
	}

	// Calculate error rate
	totalRequests := float64(ch.Metrics.TotalRequests)
	totalErrors := float64(ch.Metrics.TotalRateLimitErrors + ch.Metrics.TotalRetries)
	errorRate := totalErrors / totalRequests

	// Set the new error rate in the metrics
	ch.Metrics.ResponseCodeMetrics.ErrorRate = errorRate

	// Check if the error rate exceeds the threshold and adjust concurrency accordingly
	if errorRate > ErrorRateThreshold && len(ch.sem) > MinConcurrency {
		// Decrease concurrency
		newSize := len(ch.sem) - 1
		ch.logger.Info("Reducing request concurrency due to high error rate", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else if errorRate <= ErrorRateThreshold && len(ch.sem) < MaxConcurrency {
		// Scale up if error rate is below the threshold and concurrency is below the maximum limit
		newSize := len(ch.sem) + 1
		ch.logger.Info("Increasing request concurrency due to low error rate", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	}
}

// MonitorResponseTimeVariability calculates the standard deviation of response times
// and uses moving averages to smooth out fluctuations, adjusting concurrency accordingly.
func (ch *ConcurrencyHandler) MonitorResponseTimeVariability(responseTime time.Duration) {
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Update ResponseTimeVariability metrics
	ch.Metrics.ResponseTimeVariability.Lock.Lock()
	defer ch.Metrics.ResponseTimeVariability.Lock.Unlock()
	ch.Metrics.ResponseTimeVariability.Total += responseTime
	ch.Metrics.ResponseTimeVariability.Count++

	// Calculate average response time
	ch.Metrics.ResponseTimeVariability.Average = ch.Metrics.ResponseTimeVariability.Total / time.Duration(ch.Metrics.ResponseTimeVariability.Count)

	// Calculate variance of response times
	ch.Metrics.ResponseTimeVariability.Variance = ch.calculateVariance(ch.Metrics.ResponseTimeVariability.Average, responseTime)

	// Calculate standard deviation of response times
	stdDev := math.Sqrt(ch.Metrics.ResponseTimeVariability.Variance)

	// Adjust concurrency based on response time variability
	if stdDev > ch.Metrics.ResponseTimeVariability.StdDevThreshold && len(ch.sem) > MinConcurrency {
		newSize := len(ch.sem) - 1
		ch.logger.Info("Reducing request concurrency due to high response time variability", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else if stdDev <= ch.Metrics.ResponseTimeVariability.StdDevThreshold && len(ch.sem) < MaxConcurrency {
		// Scale up if response time variability is below the threshold and concurrency is below the maximum limit
		newSize := len(ch.sem) + 1
		ch.logger.Info("Increasing request concurrency due to low response time variability", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	}
}

// calculateVariance calculates the variance of response times.
func (ch *ConcurrencyHandler) calculateVariance(averageResponseTime time.Duration, responseTime time.Duration) float64 {
	// Convert time.Duration values to seconds
	averageSeconds := averageResponseTime.Seconds()
	responseSeconds := responseTime.Seconds()

	// Calculate variance
	variance := (float64(ch.Metrics.ResponseTimeVariability.Count-1)*math.Pow(averageSeconds-responseSeconds, 2) + ch.Metrics.ResponseTimeVariability.Variance) / float64(ch.Metrics.ResponseTimeVariability.Count)
	ch.Metrics.ResponseTimeVariability.Variance = variance
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
		ch.logger.Info("Reducing request concurrency due to high TTFB", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else if throughputMovingAverage > MaxAcceptableThroughput && len(ch.sem) < MaxConcurrency {
		newSize := len(ch.sem) + 1
		ch.logger.Info("Increasing request concurrency due to high throughput", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	}
}
