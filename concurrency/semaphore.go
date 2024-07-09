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

// AcquireConcurrencyPermit acquires a concurrency permit to manage the number of simultaneous
// operations within predefined limits. This method ensures system stability and compliance
// with concurrency policies by regulating the execution of concurrent operations.
//
// Parameters:
//   - ctx: A parent context which is used as the basis for permit acquisition. This allows
//     for proper handling of timeouts and cancellation in line with best practices.
//
// Returns:
//   - context.Context: A new context derived from the original, including a unique request ID.
//     This context is used to trace and manage operations under the acquired concurrency permit.
//   - uuid.UUID: The unique request ID generated during the permit acquisition process.
//   - error: An error object that indicates failure to acquire a permit within the allotted
//     timeout, or other system-related issues.
//
// Usage:
// This function should be used before initiating any operation that requires concurrency control.
// The returned context should be passed to subsequent operations to maintain consistency in
// concurrency tracking.
func (ch *ConcurrencyHandler) AcquireConcurrencyPermit(ctx context.Context) (context.Context, uuid.UUID, error) {
	log := ch.logger

	// Start measuring the permit acquisition time.
	tokenAcquisitionStart := time.Now()

	// Generate a unique request ID for this permit acquisition.
	requestID := uuid.New()

	// Create a new context with a specified timeout for acquiring the permit.
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel() // Ensure to free up resources by cancelling the context after use.

	select {
	case ch.sem <- struct{}{}: // permit acquisition was successful.
		// Record the time taken to acquire the permit.
		tokenAcquisitionDuration := time.Since(tokenAcquisitionStart)
		ch.trackResourceAcquisition(tokenAcquisitionDuration, requestID) // Track and log metrics.

		// Create a new context that includes the unique request ID.
		ctxWithRequestID := context.WithValue(ctx, RequestIDKey{}, requestID)
		return ctxWithRequestID, requestID, nil

	case <-ctxWithTimeout.Done(): // Timeout occurred before a permit could be acquired.
		log.Error("Failed to acquire concurrency permit", zap.Error(ctxWithTimeout.Err()))
		return ctx, requestID, ctxWithTimeout.Err()
	}
}

// trackResourceAcquisition logs and updates metrics associated with the acquisition of concurrency tokens.
// This method centralizes the logic for updating metrics and logging acquisition details, promoting code
// reusability and cleaner main logic in the permit acquisition method.
//
// Parameters:
//   - duration: The time duration it took to acquire the permit.
//   - requestID: The unique identifier for the request associated with this permit.
//
// This method locks the concurrency handler to safely update shared metrics and logs detailed
// information about the permit acquisition for debugging and monitoring purposes.
func (ch *ConcurrencyHandler) trackResourceAcquisition(duration time.Duration, requestID uuid.UUID) {
	ch.Lock()
	defer ch.Unlock()

	// Record the time taken to acquire the permit and update related metrics.
	ch.AcquisitionTimes = append(ch.AcquisitionTimes, duration)
	ch.Metrics.Lock()
	ch.Metrics.PermitWaitTime += duration
	ch.Metrics.TotalRequests++ // Increment the count of total requests handled.
	ch.Metrics.Unlock()

	// Calculate and log the current state of permit utilization.
	utilizedPermits := len(ch.sem)
	availablePermits := cap(ch.sem) - utilizedPermits
	ch.logger.Debug("Resource acquired", zap.String("RequestID", requestID.String()), zap.Duration("Duration", duration), zap.Int("UtilizedPermits", utilizedPermits), zap.Int("AvailablePermits", availablePermits))
}

// ReleaseConcurrencyPermit releases a concurrency permit back to the semaphore, making it available for other
// operations. This function is essential for maintaining the health and efficiency of the application's concurrency
// control system by ensuring that resources are properly recycled and available for use by subsequent operations.
//
// Parameters:
//   - requestID: The unique identifier for the request associated with the permit being released. This ID is used
//     for structured logging to aid in tracking and debugging permit lifecycle events.
//
// Usage:
// This method should be called as soon as a request or operation that required a concurrency permit is completed.
// It ensures that concurrency limits are adhered to and helps prevent issues such as permit leakage or semaphore saturation,
// which could lead to degraded performance or deadlock conditions.
//
// Example:
// defer concurrencyHandler.ReleaseConcurrencyPermit(requestID)
// This usage ensures that the permit is released in a deferred manner at the end of the operation, regardless of
// how the operation exits (normal completion or error path).
func (ch *ConcurrencyHandler) ReleaseConcurrencyPermit(requestID uuid.UUID) {
	// Safely remove a permit from the semaphore to make it available for other operations.
	select {
	case <-ch.sem:
		// Continue to process after successfully retrieving a permit from the semaphore.
	default:
		// Log an error if no permit was available to release, indicating a potential synchronization issue.
		ch.logger.Error("Attempted to release a non-existent concurrency permit", zap.String("RequestID", requestID.String()))
		return
	}

	ch.Lock()
	defer ch.Unlock()

	// Update metrics related to permit release.
	ch.Metrics.Lock()
	ch.Metrics.TotalRequests-- // Decrement the count of total requests handled, if applicable.
	ch.Metrics.Unlock()

	utilizedPermits := len(ch.sem)                    // Calculate tokens currently in use.
	availablePermits := cap(ch.sem) - utilizedPermits // Calculate tokens that are available for use.

	// Log the release of the concurrency permit for auditing and debugging purposes.
	ch.logger.Debug("Released concurrency permit",
		zap.String("RequestID", requestID.String()),
		zap.Int("UtilizedPermits", utilizedPermits),
		zap.Int("AvailablePermits", availablePermits),
	)
}
