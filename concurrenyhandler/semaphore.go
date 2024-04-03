// concurrencyhandler/semaphore.go
/* package provides utilities to manage concurrency control. The Concurrency Manager
ensures no more than a certain number of concurrent requests (e.g., 5 for Jamf Pro)
are sent at the same time. This is managed using a semaphore */
package concurrencyhandler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

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
func (ch *ConcurrencyHandler) AcquireConcurrencyToken(ctx context.Context) (context.Context, error) {
	log := ch.logger

	// Measure the token acquisition start time
	tokenAcquisitionStart := time.Now()

	// Create a new context with a timeout for acquiring the concurrency token
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Generate a unique request ID for this acquisition
	requestID := uuid.New()

	// Attempt to acquire a token from the semaphore within the given context timeout
	select {
	case ch.sem <- struct{}{}: // Successfully acquired a token
		// Calculate the duration it took to acquire the token and record it
		tokenAcquisitionDuration := time.Since(tokenAcquisitionStart)
		ch.lock.Lock()
		ch.AcquisitionTimes = append(ch.AcquisitionTimes, tokenAcquisitionDuration)
		ch.PerfMetrics.lock.Lock()
		ch.PerfMetrics.TokenWaitTime += tokenAcquisitionDuration
		ch.PerfMetrics.lock.Unlock()
		ch.lock.Unlock()

		// Add the acquired request ID to the context for use in subsequent operations
		ctxWithRequestID := context.WithValue(ctx, requestIDKey{}, requestID)

		// Return the updated context and nil error to indicate success
		return ctxWithRequestID, nil

	case <-ctxWithTimeout.Done(): // Failed to acquire a token within the timeout
		log.Error("Failed to acquire concurrency token", zap.Error(ctxWithTimeout.Err()))
		return nil, ctxWithTimeout.Err()
	}
}

// Acquire attempts to get a token to allow an HTTP request to proceed.
// It blocks until a token is available or the context expires.
// Returns a unique request ID upon successful acquisition.
func (ch *ConcurrencyHandler) Acquire(ctx context.Context) (uuid.UUID, error) {
	requestID := uuid.New()
	startTime := time.Now()

	select {
	case ch.sem <- struct{}{}:
		acquisitionTime := time.Since(startTime)
		ch.lock.Lock()
		ch.AcquisitionTimes = append(ch.AcquisitionTimes, acquisitionTime)
		ch.lock.Unlock()
		ch.lastTokenAcquisitionTime = time.Now()

		utilizedTokens := len(ch.sem)
		availableTokens := cap(ch.sem) - utilizedTokens
		ch.logger.Debug("Acquired concurrency token",
			zap.String("ConcurrencyTokenID", requestID.String()),
			zap.Duration("AcquisitionTime", acquisitionTime),
			zap.Int("UtilizedTokens", utilizedTokens),
			zap.Int("AvailableTokens", availableTokens),
		)
		return requestID, nil

	case <-ctx.Done():
		ch.logger.Warn("Failed to acquire concurrency token, context done",
			zap.String("ConcurrencyTokenID", requestID.String()),
			zap.Error(ctx.Err()),
		)
		return requestID, ctx.Err()
	}
}

// Release returns a token back to the pool, allowing other requests to proceed.
// It uses the provided requestID for logging and debugging purposes.
func (ch *ConcurrencyHandler) Release(requestID uuid.UUID) {
	<-ch.sem                                        // Release a token back to the semaphore
	utilizedTokens := len(ch.sem)                   // Tokens currently in use
	availableTokens := cap(ch.sem) - utilizedTokens // Tokens available for use

	// Using zap fields for structured logging in debug mode
	ch.logger.Debug("Released concurrency token",
		zap.String("ConcurrencyTokenID", requestID.String()),
		zap.Int("UtilizedTokens", utilizedTokens),
		zap.Int("AvailableTokens", availableTokens),
	)
}
