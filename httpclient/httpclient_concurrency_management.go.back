// http_concurrency_management.go
// package httpclient provides utilities to manage HTTP client interactions, including concurrency control.
// The Concurrency Manager ensures no more than a certain number of concurrent requests
// (e.g., 5 for Jamf Pro) are sent at the same time. This is managed using a semaphore
package httpclient

import (
	"context"
	"sync"
	"time"

	"github.com/deploymenttheory/go-api-http-client/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Constants and Data Structures:

const (
	MaxConcurrency     = 10              // Maximum allowed concurrent requests
	MinConcurrency     = 1               // Minimum allowed concurrent requests
	EvaluationInterval = 1 * time.Minute // Time interval for evaluating metrics and adjusting concurrency
)

// ConcurrencyManager controls the number of concurrent HTTP requests.
type ConcurrencyManager struct {
	sem                      chan struct{}
	logger                   logger.Logger
	debugMode                bool
	AcquisitionTimes         []time.Duration
	lock                     sync.Mutex
	lastTokenAcquisitionTime time.Time
}

// requestIDKey is type used as a key for storing and retrieving
// request-specific identifiers from a context.Context object. This private
// type ensures that the key is distinct and prevents accidental value
// retrieval or conflicts with other context keys. The value associated
// with this key in a context is typically a UUID that uniquely identifies
// a request being processed by the ConcurrencyManager, allowing for
// fine-grained control and tracking of concurrent HTTP requests.
type requestIDKey struct{}

// Functions:

// NewConcurrencyManager initializes a new ConcurrencyManager with the given
// concurrency limit, logger, and debug mode. The ConcurrencyManager ensures
// no more than a certain number of concurrent requests are made.
// It uses a semaphore to control concurrency.
func NewConcurrencyManager(limit int, logger logger.Logger, debugMode bool) *ConcurrencyManager {
	return &ConcurrencyManager{
		sem:              make(chan struct{}, limit),
		logger:           logger,
		debugMode:        debugMode,
		AcquisitionTimes: []time.Duration{},
	}
}

// AcquireConcurrencyToken attempts to acquire a token from the ConcurrencyManager
// to manage the number of concurrent requests. This function is designed to ensure
// that the HTTP client adheres to predefined concurrency limits, preventing an
// excessive number of simultaneous requests. It creates a new context with a timeout
// to avoid indefinite blocking in case the concurrency limit is reached.
// Upon successfully acquiring a token, it records the time taken to acquire the
// token and updates performance metrics accordingly. The function then adds the
// acquired request ID to the context, which can be used for tracking and managing
// individual requests.
//
// Parameters:
// - ctx: The parent context from which the new context with timeout will be derived.
// This allows for proper request cancellation and timeout handling.
//
// Returns:
// - A new context containing the acquired request ID, which should be passed to
// subsequent operations requiring concurrency control.
//
// - An error if the token could not be acquired within the timeout period or due to
// any other issues encountered by the ConcurrencyManager.
//
// Usage:
// This function should be called before making an HTTP request that needs to be
// controlled for concurrency. The returned context should be used for the HTTP
// request to ensure it is associated with the acquired concurrency token.
func (c *Client) AcquireConcurrencyToken(ctx context.Context) (context.Context, error) {
	log := c.Logger

	// Measure the token acquisition start time
	tokenAcquisitionStart := time.Now()

	// Create a new context with a timeout for acquiring the concurrency token
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	requestID, err := c.ConcurrencyMgr.Acquire(ctxWithTimeout)
	if err != nil {
		log.Error("Failed to acquire concurrency token", zap.Error(err))
		return nil, err
	}

	// Calculate the duration it took to acquire the token
	tokenAcquisitionDuration := time.Since(tokenAcquisitionStart)

	// Lock the mutex before updating the performance metrics
	c.PerfMetrics.lock.Lock()
	c.PerfMetrics.TokenWaitTime += tokenAcquisitionDuration
	c.PerfMetrics.lock.Unlock()

	// Add the acquired request ID to the context for use in subsequent operations
	ctxWithRequestID := context.WithValue(ctx, requestIDKey{}, requestID)

	// Return the updated context and nil error to indicate success
	return ctxWithRequestID, nil
}

// updatePerformanceMetrics updates the client's performance metrics by recording the duration
// of the HTTP request and incrementing the total request count. This function is thread-safe
// and uses a mutex to synchronize updates to the performance metrics.
//
// Parameters:
// - duration: The time duration it took for an HTTP request to complete.
//
// This function should be called after each HTTP request to keep track of the client's
// performance over time.
func (c *Client) updatePerformanceMetrics(duration time.Duration) {
	c.PerfMetrics.lock.Lock()
	defer c.PerfMetrics.lock.Unlock()
	c.PerfMetrics.TotalResponseTime += duration
	c.PerfMetrics.TotalRequests++
}

// Min returns the smaller of the two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

//------ Core Concurrency Functions:

// Acquire attempts to get a token to allow an HTTP request to proceed.
// It blocks until a token is available or the context expires.
// Returns a unique request ID upon successful acquisition.
func (c *ConcurrencyManager) Acquire(ctx context.Context) (uuid.UUID, error) {
	requestID := uuid.New()
	startTime := time.Now()

	select {
	case c.sem <- struct{}{}:
		acquisitionTime := time.Since(startTime)
		c.lock.Lock()
		c.AcquisitionTimes = append(c.AcquisitionTimes, acquisitionTime)
		c.lock.Unlock()
		c.lastTokenAcquisitionTime = time.Now()

		utilizedTokens := len(c.sem)
		availableTokens := cap(c.sem) - utilizedTokens
		c.logger.Debug("Acquired concurrency token",
			zap.String("ConcurrencyTokenID", requestID.String()),
			zap.Duration("AcquisitionTime", acquisitionTime),
			zap.Int("UtilizedTokens", utilizedTokens),
			zap.Int("AvailableTokens", availableTokens),
		)
		return requestID, nil

	case <-ctx.Done():
		c.logger.Warn("Failed to acquire concurrency token, context done",
			zap.String("ConcurrencyTokenID", requestID.String()),
			zap.Error(ctx.Err()),
		)
		return requestID, ctx.Err()
	}
}

// Release returns a token back to the pool, allowing other requests to proceed.
// It uses the provided requestID for logging and debugging purposes.
func (c *ConcurrencyManager) Release(requestID uuid.UUID) {
	<-c.sem // Release a token back to the semaphore
	if c.debugMode {
		utilizedTokens := len(c.sem)                   // Tokens currently in use
		availableTokens := cap(c.sem) - utilizedTokens // Tokens available for use

		// Using zap fields for structured logging in debug mode
		c.logger.Debug("Released concurrency token",
			zap.String("ConcurrencyTokenID", requestID.String()),
			zap.Int("UtilizedTokens", utilizedTokens),
			zap.Int("AvailableTokens", availableTokens),
		)
	}
}

//------ Metric-related Functions:

// AverageAcquisitionTime computes the average time taken to acquire a token
// from the semaphore. It helps in understanding the contention for tokens
// and can be used to adjust concurrency limits.
func (c *ConcurrencyManager) AverageAcquisitionTime() time.Duration {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.AcquisitionTimes) == 0 {
		return 0
	}

	totalTime := time.Duration(0)
	for _, t := range c.AcquisitionTimes {
		totalTime += t
	}
	return totalTime / time.Duration(len(c.AcquisitionTimes))
}

// HistoricalAverageAcquisitionTime computes the average time taken to acquire
// a token from the semaphore over a historical period (e.g., the last 5 minutes).
// It helps in understanding the historical contention for tokens and can be used
// to adjust concurrency limits.
func (c *ConcurrencyManager) HistoricalAverageAcquisitionTime() time.Duration {
	c.lock.Lock()
	defer c.lock.Unlock()

	// For simplicity, let's say we store the last 5 minutes of acquisition times.
	// This means if EvaluationInterval is 1 minute, we consider the last 5 data points.
	historicalCount := 5
	if len(c.AcquisitionTimes) < historicalCount {
		return c.AverageAcquisitionTime() // If not enough historical data, return the overall average
	}

	totalTime := time.Duration(0)
	for _, t := range c.AcquisitionTimes[len(c.AcquisitionTimes)-historicalCount:] {
		totalTime += t
	}
	return totalTime / time.Duration(historicalCount)
}

//------ Concurrency Adjustment Functions:

// AdjustConcurrencyLimit dynamically modifies the maximum concurrency limit
// based on the newLimit provided. This function helps in adjusting the concurrency
// limit in real-time based on observed system performance and other metrics. It
// transfers the tokens from the old semaphore to the new one, ensuring that there's
// no loss of tokens during the transition.
func (c *ConcurrencyManager) AdjustConcurrencyLimit(newLimit int) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if newLimit <= 0 {
		return // Avoid setting a non-positive limit
	}

	// Create a new semaphore with the desired limit
	newSem := make(chan struct{}, newLimit)

	// Transfer tokens from the old semaphore to the new one
	for i := 0; i < len(c.sem) && i < newLimit; i++ {
		newSem <- struct{}{}
	}

	c.sem = newSem
}

// AdjustConcurrencyBasedOnMetrics evaluates the current metrics and adjusts the
// concurrency limit if required. It checks metrics like average token acquisition
// time and decides on a new concurrency limit. The method ensures that the new
// limit respects the minimum and maximum allowed concurrency bounds.
func (c *Client) AdjustConcurrencyBasedOnMetrics() {
	log := c.Logger
	// Get average acquisition time
	avgAcquisitionTime := c.ConcurrencyMgr.AverageAcquisitionTime()

	// Get current concurrency limit
	currentLimit := cap(c.ConcurrencyMgr.sem)

	// Get historical average acquisition time (e.g., over the last 5 minutes)
	historicalAvgAcquisitionTime := c.ConcurrencyMgr.HistoricalAverageAcquisitionTime()

	// Decide on new limit based on metrics
	newLimit := currentLimit
	if avgAcquisitionTime > time.Duration(float64(historicalAvgAcquisitionTime)*1.2) { // 20% increase in acquisition time
		newLimit = currentLimit - 2 // decrease concurrency more aggressively
	} else if avgAcquisitionTime < time.Duration(float64(historicalAvgAcquisitionTime)*0.8) { // 20% decrease in acquisition time
		newLimit = currentLimit + 2 // increase concurrency more aggressively
	} else if avgAcquisitionTime > historicalAvgAcquisitionTime {
		newLimit = currentLimit - 1 // decrease concurrency conservatively
	} else if avgAcquisitionTime < historicalAvgAcquisitionTime {
		newLimit = currentLimit + 1 // increase concurrency conservatively
	}

	// Ensure newLimit is within safety bounds
	if newLimit > MaxConcurrency {
		newLimit = MaxConcurrency
	} else if newLimit < MinConcurrency {
		newLimit = MinConcurrency
	}

	// Adjust concurrency if new limit is different from current
	if newLimit != currentLimit {
		c.ConcurrencyMgr.AdjustConcurrencyLimit(newLimit)

		log.Debug("Adjusted concurrency",
			zap.Int("OldLimit", currentLimit),
			zap.Int("NewLimit", newLimit),
			zap.String("Reason", "Based on average acquisition time"),
			zap.Duration("AverageAcquisitionTime", avgAcquisitionTime),
			zap.Duration("HistoricalAverageAcquisitionTime", historicalAvgAcquisitionTime),
		)
	}
}

// EvaluateMetricsAndAdjustConcurrency evaluates the performance metrics and makes necessary
// adjustments to the concurrency limit. The method assesses the average response time
// and adjusts the concurrency based on how it compares to the historical average acquisition time.
// If the average response time has significantly increased compared to the historical average,
// the concurrency limit is decreased, and vice versa. The method ensures that the concurrency
// limit remains within the bounds defined by the system's best practices.
func (c *Client) EvaluateMetricsAndAdjustConcurrency() {
	c.PerfMetrics.lock.Lock()
	averageResponseTime := c.PerfMetrics.TotalResponseTime / time.Duration(c.PerfMetrics.TotalRequests)
	c.PerfMetrics.lock.Unlock()

	historicalAverageAcquisitionTime := c.ConcurrencyMgr.HistoricalAverageAcquisitionTime()

	if averageResponseTime > time.Duration(float64(historicalAverageAcquisitionTime)*1.2) {
		// Decrease concurrency
		currentLimit := cap(c.ConcurrencyMgr.sem)
		newLimit := currentLimit - 1
		if newLimit < MinConcurrency {
			newLimit = MinConcurrency
		}
		c.ConcurrencyMgr.AdjustConcurrencyLimit(newLimit)
	} else if averageResponseTime < time.Duration(float64(historicalAverageAcquisitionTime)*0.8) {
		// Increase concurrency
		currentLimit := cap(c.ConcurrencyMgr.sem)
		newLimit := currentLimit + 1
		if newLimit > MaxConcurrency || newLimit > 5 {
			newLimit = Min(currentLimit, 5)
		}
		c.ConcurrencyMgr.AdjustConcurrencyLimit(newLimit)
	}
}

//------ Concurrency Monitoring Functions:

// StartMetricEvaluation continuously monitors the client's interactions with the API and adjusts the concurrency limits dynamically.
// The function evaluates metrics at regular intervals to detect burst activity patterns.
// If a burst activity is detected (e.g., many requests in a short period), the evaluation interval is reduced for more frequent checks.
// Otherwise, it reverts to a default interval for regular checks.
// After each evaluation, the function calls EvaluateMetricsAndAdjustConcurrency to potentially adjust the concurrency based on observed metrics.
//
// The evaluation process works as follows:
// 1. Sleep for the defined evaluation interval.
// 2. Check if there's a burst in activity using the isBurstActivity method.
// 3. If a burst is detected, the evaluation interval is shortened to more frequently monitor and adjust the concurrency.
// 4. If no burst is detected, it maintains the default evaluation interval.
// 5. It then evaluates the metrics and adjusts the concurrency accordingly.
func (c *Client) StartMetricEvaluation() {
	evalInterval := 5 * time.Minute // Initial interval

	for {
		time.Sleep(evalInterval)

		if c.isBurstActivity() {
			evalInterval = 1 * time.Minute
		} else {
			evalInterval = 5 * time.Minute
		}

		c.EvaluateMetricsAndAdjustConcurrency()
	}
}

func (c *Client) isBurstActivity() bool {
	// If the last token was acquired less than 2 minutes ago, consider it a burst
	return time.Since(c.ConcurrencyMgr.lastTokenAcquisitionTime) < 2*time.Minute
}

// StartConcurrencyAdjustment launches a periodic checker that evaluates current metrics and adjusts concurrency limits if needed.
// It uses a ticker to periodically trigger the adjustment logic.
func (c *Client) StartConcurrencyAdjustment() {
	ticker := time.NewTicker(EvaluationInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.AdjustConcurrencyBasedOnMetrics()
	}
}

// Returns the average Acquisition Time to get a token from the semaphore
func (c *Client) AverageAcquisitionTime() time.Duration {
	// Assuming ConcurrencyMgr has a method to get this metric
	return c.ConcurrencyMgr.AverageAcquisitionTime()
}

func (c *Client) HistoricalAverageAcquisitionTime() time.Duration {
	// Assuming ConcurrencyMgr has a method to get this metric
	return c.ConcurrencyMgr.HistoricalAverageAcquisitionTime()
}

// Returns performance metrics from the http client
func (c *Client) GetPerformanceMetrics() *PerformanceMetrics {
	return &c.PerfMetrics
}
