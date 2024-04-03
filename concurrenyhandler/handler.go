// concurrencyhandler/handler.go
package concurrencyhandler

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
	PerfMetrics              *PerformanceMetrics
}

// PerformanceMetrics captures various metrics related to the client's
// interactions with the API.
type PerformanceMetrics struct {
	TotalRequests        int64
	TotalRetries         int64
	TotalRateLimitErrors int64
	TotalResponseTime    time.Duration
	TokenWaitTime        time.Duration
	lock                 sync.Mutex // Protects performance metrics fields
}

// NewConcurrencyManager initializes a new ConcurrencyManager with the given
// concurrency limit, logger, and perf metrics. The ConcurrencyManager ensures
// no more than a certain number of concurrent requests are made.
// It uses a semaphore to control concurrency.
func NewConcurrencyHandler(limit int, logger logger.Logger, perfMetrics *PerformanceMetrics) *ConcurrencyHandler {
	return &ConcurrencyHandler{
		sem:              make(chan struct{}, limit),
		logger:           logger,
		AcquisitionTimes: []time.Duration{},
		PerfMetrics:      perfMetrics,
	}
}

// requestIDKey is type used as a key for storing and retrieving
// request-specific identifiers from a context.Context object. This private
// type ensures that the key is distinct and prevents accidental value
// retrieval or conflicts with other context keys. The value associated
// with this key in a context is typically a UUID that uniquely identifies
// a request being processed by the ConcurrencyManager, allowing for
// fine-grained control and tracking of concurrent HTTP requests.
type requestIDKey struct{}
