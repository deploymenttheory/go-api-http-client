// concurrency/handler.go
package concurrency

import (
	"sync"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
)

// Constants and Data Structures:
const (
	MaxConcurrency     = 10              // Maximum allowed concurrent requests
	MinConcurrency     = 1               // Minimum allowed concurrent requests
	EvaluationInterval = 1 * time.Minute // Time interval for evaluating metrics and adjusting concurrency
)

// ConcurrencyHandler controls the number of concurrent HTTP requests.
type ConcurrencyHandler struct {
	sem                      chan struct{}
	logger                   logger.Logger
	AcquisitionTimes         []time.Duration
	lock                     sync.Mutex
	lastTokenAcquisitionTime time.Time
	Metrics                  *ConcurrencyMetrics
}

// ConcurrencyMetrics captures various metrics related to managing concurrency for the client's interactions with the API.// ConcurrencyMetrics captures various metrics related to managing concurrency for the client's interactions with the API.
type ConcurrencyMetrics struct {
	TotalRequests        int64         // Total number of requests made
	TotalRetries         int64         // Total number of retry attempts
	TotalRateLimitErrors int64         // Total number of rate limit errors encountered
	TotalResponseTime    time.Duration // Total response time for all requests
	AverageResponseTime  time.Duration // Average response time across all requests
	ErrorRate            float64       // Error rate calculated as (TotalRateLimitErrors + 5xxErrors) / TotalRequests
	TokenWaitTime        time.Duration // Total time spent waiting for tokens
	TTFB                 struct {      // Metrics related to Time to First Byte (TTFB)
		Total time.Duration // Total Time to First Byte (TTFB) for all requests
		Count int64         // Count of requests used for calculating TTFB
		Lock  sync.Mutex    // Lock for TTFB metrics
	}
	Throughput struct { // Metrics related to network throughput
		Total float64    // Total network throughput for all requests
		Count int64      // Count of requests used for calculating throughput
		Lock  sync.Mutex // Lock for throughput metrics
	}
	Variance      float64    // Variance of response times
	ResponseCount int64      // Count of responses used for calculating response time variability
	Lock          sync.Mutex // Lock for overall metrics fields
}

// NewConcurrencyHandler initializes a new ConcurrencyHandler with the given
// concurrency limit, logger, and concurrency metrics. The ConcurrencyHandler ensures
// no more than a certain number of concurrent requests are made.
// It uses a semaphore to control concurrency.
func NewConcurrencyHandler(limit int, logger logger.Logger, metrics *ConcurrencyMetrics) *ConcurrencyHandler {
	return &ConcurrencyHandler{
		sem:              make(chan struct{}, limit),
		logger:           logger,
		AcquisitionTimes: []time.Duration{},
		Metrics:          metrics,
	}
}

// RequestIDKey is type used as a key for storing and retrieving
// request-specific identifiers from a context.Context object. This private
// type ensures that the key is distinct and prevents accidental value
// retrieval or conflicts with other context keys. The value associated
// with this key in a context is typically a UUID that uniquely identifies
// a request being processed by the ConcurrencyManager, allowing for
// fine-grained control and tracking of concurrent HTTP requests.
type RequestIDKey struct{}
