// concurrency/adjust_concurrency.go

package concurrency

import (
	"time"

	"go.uber.org/zap"
)

// AdjustConcurrency dynamically adjusts the number of concurrent requests allowed based on the current
// metrics tracked by the ConcurrencyHandler. This function assesses the average response time and error
// rate to decide whether to increase or decrease the concurrency limits.
//
// If the average response time exceeds one second and the current concurrency is greater than the minimum
// limit, the concurrency level is decreased to help alleviate potential load on the server or network.
// Conversely, if the error rate is below 0.05 and the current concurrency is below the maximum limit,
// the concurrency level is increased to potentially improve throughput.
//
// This function locks the metrics to ensure thread safety and prevent race conditions during the read
// and update operations. The actual adjustment of the semaphore's size is delegated to the ResizeSemaphore
// function.
func (ch *ConcurrencyHandler) AdjustConcurrency() {
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Example logic based on simplified conditions
	if ch.Metrics.AverageResponseTime > time.Second && len(ch.sem) > MinConcurrency {
		newSize := len(ch.sem) - 1
		ch.logger.Info("Reducing concurrency due to high response time", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else if ch.Metrics.ErrorRate < 0.05 && len(ch.sem) < MaxConcurrency {
		newSize := len(ch.sem) + 1
		ch.logger.Info("Increasing concurrency due to low error rates", zap.Int("NewSize", newSize))
		ch.ResizeSemaphore(newSize)
	}
}

// ResizeSemaphore adjusts the size of the semaphore used to control concurrency. This method creates a new
// semaphore with the specified new size and closes the old semaphore to ensure that no further tokens can
// be acquired from it. This approach helps manage the transition from the old concurrency level to the new one
// without affecting ongoing operations significantly.
//
// Parameters:
//   - newSize: The new size for the semaphore, representing the updated limit on concurrent requests.
//
// This function should be called from within synchronization contexts, such as AdjustConcurrency, to avoid
// race conditions and ensure that changes to the semaphore are consistent with the observed metrics.
func (ch *ConcurrencyHandler) ResizeSemaphore(newSize int) {
	newSem := make(chan struct{}, newSize)

	// Transfer tokens from the old semaphore to the new one.
	for {
		select {
		case token := <-ch.sem:
			select {
			case newSem <- token:
				// Token transferred to new semaphore.
			default:
				// New semaphore is full, put token back to the old one to allow ongoing operations to complete.
				ch.sem <- token
			}
		default:
			// No more tokens to transfer.
			close(ch.sem)
			ch.sem = newSem
			return
		}
	}
}
