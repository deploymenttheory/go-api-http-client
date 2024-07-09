// concurrency/handler.go
package concurrency

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConcurrencyHandler controls the number of concurrent HTTP requests.
type ConcurrencyHandler struct {
	sem                      chan struct{}
	logger                   *zap.SugaredLogger
	AcquisitionTimes         []time.Duration
	lastTokenAcquisitionTime time.Time
	Metrics                  *ConcurrencyMetrics
	sync.Mutex
}

// ConcurrencyMetrics captures various metrics related to managing concurrency for the client's interactions with the API.
type ConcurrencyMetrics struct {
	TotalRequests        int64
	TotalRetries         int64
	TotalRateLimitErrors int64
	PermitWaitTime       time.Duration
	sync.Mutex
	TTFB struct {
		Total time.Duration
		Count int64
		sync.Mutex
	}
	Throughput struct {
		Total float64
		Count int64
		sync.Mutex
	}
	ResponseTimeVariability struct {
		Total    time.Duration
		Average  time.Duration
		Variance float64
		Count    int64
		sync.Mutex
		StdDevThreshold        float64
		DebounceScaleDownCount int
		DebounceScaleUpCount   int
	}
	ResponseCodeMetrics struct {
		ErrorRate float64
		sync.Mutex
	}
}

// NewConcurrencyHandler initializes a new ConcurrencyHandler with the given
// concurrency limit, logger, and concurrency metrics. The ConcurrencyHandler ensures
// no more than a certain number of concurrent requests are made.
// It uses a semaphore to control concurrency.
func NewConcurrencyHandler(limit int, logger *zap.SugaredLogger, metrics *ConcurrencyMetrics) *ConcurrencyHandler {
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
