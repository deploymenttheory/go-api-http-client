// concurrency/const.go
package concurrency

import "time"

const (
	// Concurrency constants define parameters related to managing concurrent requests.

	// MaxConcurrency represents the maximum number of concurrent requests the system is designed to handle safely.
	MaxConcurrency = 10

	// MinConcurrency is the minimum number of concurrent requests that the system will maintain,
	// even under low traffic conditions or when scaling down due to low resource utilization.
	MinConcurrency = 1

	// EvaluationInterval specifies the frequency at which the system evaluates its performance metrics
	// to make decisions about scaling concurrency up or down.
	EvaluationInterval = 1 * time.Minute

	// Threshold constants define critical operational metrics for adjusting concurrency.

	// MaxAcceptableTTFB (Time to First Byte) is the threshold for the longest acceptable delay
	// between making a request and receiving the first byte of data in the response. If response
	// times exceed this threshold, it indicates potential performance issues, and the system may
	// scale down concurrency to reduce load on the server.
	MaxAcceptableTTFB = 300 * time.Millisecond

	// MaxAcceptableThroughput is the threshold for the maximum network data transfer rate. If the
	// system's throughput exceeds this value, it may be an indicator of high traffic demanding
	// significant bandwidth, which could warrant a scale-up in concurrency to maintain performance.
	MaxAcceptableThroughput = 5 * 1024 * 1024 // 5 MBps

	// MaxAcceptableResponseTimeVariability is the threshold for the maximum allowed variability or
	// fluctuations in response times. A high variability often indicates an unstable system, which
	// could trigger a scale-down to allow the system to stabilize.
	MaxAcceptableResponseTimeVariability = 500 * time.Millisecond

	// ErrorRateThreshold is the maximum acceptable rate of error responses (such as rate-limit errors
	// and 5xx server errors) compared to the total number of requests. Exceeding this threshold suggests
	// the system is encountering issues that may be alleviated by scaling down concurrency.
	ErrorRateThreshold = 0.1 // 10% error rate

	// RateLimitCriticalThreshold defines the number of available rate limit slots considered critical.
	// Falling at or below this threshold suggests the system is close to hitting the rate limit enforced
	// by the external service, and it should scale down to prevent rate-limiting errors.
	RateLimitCriticalThreshold = 5

	// ErrorResponseThreshold is the threshold for the error rate that, once exceeded, indicates the system
	// should consider scaling down. It is a ratio of the number of error responses to the total number of
	// requests, reflecting the health of the interaction with the external system.
	ErrorResponseThreshold = 0.2 // 20% error rate

	// ResponseTimeCriticalThreshold is the duration beyond which the response time is considered critically
	// high. If response times exceed this threshold, it could signal that the system or the external service
	// is under heavy load and may benefit from scaling down concurrency to alleviate pressure.
	ResponseTimeCriticalThreshold = 2 * time.Second
)
