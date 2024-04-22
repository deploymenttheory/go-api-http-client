// concurrency/const.go
package concurrency

import "time"

const (
	// Concurrency constants define parameters related to managing concurrent requests.

	// MaxConcurrency represents the maximum allowed concurrent requests.
	MaxConcurrency = 10

	// MinConcurrency represents the minimum allowed concurrent requests.
	MinConcurrency = 1

	// EvaluationInterval defines the time interval for evaluating metrics and adjusting concurrency.
	EvaluationInterval = 1 * time.Minute

	// Threshold constants define thresholds for adjusting concurrency based on various metrics.

	// MaxAcceptableTTFB represents the maximum acceptable Time to First Byte (TTFB) in milliseconds.
	// TTFB is the time taken for the server to start sending the first byte of data in response to a request.
	// Adjustments in concurrency will be made if the TTFB exceeds this threshold.
	MaxAcceptableTTFB = 300 * time.Millisecond

	// MaxAcceptableThroughput represents the maximum acceptable network throughput in bytes per second.
	// Throughput is the amount of data transferred over the network within a specific time interval.
	// Adjustments in concurrency will be made if the network throughput exceeds this threshold.
	MaxAcceptableThroughput = 5 * 1024 * 1024

	// MaxAcceptableResponseTimeVariability represents the maximum acceptable variability in response times.
	// It is used as a threshold to dynamically adjust concurrency based on fluctuations in response times.
	MaxAcceptableResponseTimeVariability = 500 * time.Millisecond

	// ErrorRateThreshold represents the threshold for error rate above which concurrency will be adjusted.
	// Error rate is calculated as (TotalRateLimitErrors + 5xxErrors) / TotalRequests.
	// Adjustments in concurrency will be made if the error rate exceeds this threshold. A threshold of 0.1 (or 10%) is common.
	ErrorRateThreshold = 0.1
)
