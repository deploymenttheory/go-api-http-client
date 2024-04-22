// concurrency/semaphore.go
/* package provides utilities to manage concurrency control. The Concurrency Manager
ensures no more than a certain number of concurrent requests (e.g., 5 for Jamf Pro)
are sent at the same time. This is managed using a semaphore */
package concurrency

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AcquireConcurrencyToken acquires a concurrency token to regulate the number of concurrent
// operations within predefined limits, ensuring system stability and adherence to concurrency policies.
// This function initiates a token acquisition process that involves generating a unique request ID
// for tracking purposes and attempting to acquire a token within a specified timeout to prevent
// indefinite blocking. Successful acquisition updates performance metrics and associates the
// unique request ID with the provided context for enhanced traceability and management of
// concurrent requests.
//
// Parameters:
//   - ctx: A parent context used as the basis for the token acquisition attempt, facilitating
//     appropriate cancellation and timeout handling in line with best practices for concurrency control.
//
// Returns:
//   - context.Context: A derived context that includes the unique request ID, offering a mechanism
//     for associating subsequent operations with the acquired concurrency token and facilitating
//     effective request tracking and management.
//   - uuid.UUID: The unique request ID generated as part of the token acquisition process, serving
//     as an identifier for the acquired token and enabling detailed tracking and auditing of
//     concurrent operations.
//   - error: An error that signals failure to acquire a concurrency token within the allotted timeout,
//     or due to other encountered issues, ensuring that potential problems in concurrency control
//     are surfaced and can be addressed.
//
// Usage:
// This function is a critical component of concurrency control and should be invoked prior to
// executing operations that require regulation of concurrency. The returned context, enhanced
// with the unique request ID, should be utilized in the execution of these operations to maintain
// consistency in tracking and managing concurrency tokens. The unique request ID also facilitates
// detailed auditing and troubleshooting of the concurrency control mechanism.
//
// Example:
// ctx, requestID, err := concurrencyHandler.AcquireConcurrencyToken(context.Background())
//
//	if err != nil {
//	    // Handle token acquisition failure
//	}
//
// defer concurrencyHandler.Release(requestID)
// // Proceed with the operation using the modified context
func (ch *ConcurrencyHandler) AcquireConcurrencyToken(ctx context.Context) (context.Context, uuid.UUID, error) {
	log := ch.logger

	// Measure the token acquisition start time
	tokenAcquisitionStart := time.Now()

	// Generate a unique request ID for this acquisition
	requestID := uuid.New()

	// Create a new context with a timeout for acquiring the concurrency token
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Attempt to acquire a token from the semaphore within the given context timeout
	select {
	case ch.sem <- struct{}{}: // Successfully acquired a token
		// Calculate the duration it took to acquire the token and record it
		tokenAcquisitionDuration := time.Since(tokenAcquisitionStart)
		ch.lock.Lock()
		ch.AcquisitionTimes = append(ch.AcquisitionTimes, tokenAcquisitionDuration)
		ch.Metrics.Lock.Lock()
		ch.Metrics.TokenWaitTime += tokenAcquisitionDuration
		ch.Metrics.TotalRequests++ // Increment total requests count
		ch.Metrics.Lock.Unlock()
		ch.lock.Unlock()

		// Logging the acquisition
		utilizedTokens := len(ch.sem)
		availableTokens := cap(ch.sem) - utilizedTokens
		log.Debug("Acquired concurrency token", zap.String("RequestID", requestID.String()), zap.Duration("AcquisitionTime", tokenAcquisitionDuration), zap.Int("UtilizedTokens", utilizedTokens), zap.Int("AvailableTokens", availableTokens))

		// Add the acquired request ID to the context for use in subsequent operations
		ctxWithRequestID := context.WithValue(ctx, RequestIDKey{}, requestID)

		// Return the updated context, the request ID, and nil error to indicate success
		return ctxWithRequestID, requestID, nil

	case <-ctxWithTimeout.Done(): // Failed to acquire a token within the timeout
		log.Error("Failed to acquire concurrency token", zap.Error(ctxWithTimeout.Err()))
		return ctx, requestID, ctxWithTimeout.Err()
	}
}

// ReleaseConcurrencyToken returns a token back to the semaphore pool, allowing other
// operations to proceed. It uses the provided requestID for structured logging,
// aiding in tracking and debugging the release of concurrency tokens.
func (ch *ConcurrencyHandler) ReleaseConcurrencyToken(requestID uuid.UUID) {
	<-ch.sem // Release a token back to the semaphore

	ch.lock.Lock()
	defer ch.lock.Unlock()

	// Update the list of acquisition times by removing the time related to the released token
	// This step is optional and depends on whether you track acquisition times per token or not

	utilizedTokens := len(ch.sem)                   // Tokens currently in use
	availableTokens := cap(ch.sem) - utilizedTokens // Tokens available for use

	// Log the release of the concurrency token for auditing and debugging purposes
	ch.logger.Debug("Released concurrency token",
		zap.String("RequestID", requestID.String()),
		zap.Int("UtilizedTokens", utilizedTokens),
		zap.Int("AvailableTokens", availableTokens),
	)
}
