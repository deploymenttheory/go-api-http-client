// concurrency/const.go
package concurrency

import "time"

const (
	// MaxConcurrency defines the upper limit of concurrent requests the system can handle.
	MaxConcurrency = 10

	// MinConcurrency defines the lower limit of concurrent requests the system will maintain.
	MinConcurrency = 1

	// EvaluationInterval specifies the frequency at which the system evaluates its performance metrics
	// to decide if concurrency adjustments are needed.
	EvaluationInterval = 1 * time.Minute

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

	// Weight assigned to each metric feedback type
	WeightRateLimit     = 0.5 // Weight for rate limit feedback, less if not all APIs provide this data
	WeightResponseCodes = 1.0 // Weight for server response codes
	WeightResponseTime  = 1.5 // Higher weight for response time variability

	// Thresholds for semaphore scaling actions
	ThresholdScaleDown = -1.5 // Threshold to decide scaling down
	ThresholdScaleUp   = 1.5  // Threshold to decide scaling up
)
