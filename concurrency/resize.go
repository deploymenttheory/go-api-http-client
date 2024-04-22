// concurrency/resize.go
package concurrency

import (
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
)

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
// func (ch *ConcurrencyHandler) ResizeSemaphore(newSize int) {
// 	newSem := make(chan struct{}, newSize)

//		// Transfer tokens from the old semaphore to the new one.
//		for {
//			select {
//			case token := <-ch.sem:
//				select {
//				case newSem <- token:
//					// Token transferred to new semaphore.
//				default:
//					// New semaphore is full, put token back to the old one to allow ongoing operations to complete.
//					ch.sem <- token
//				}
//			default:
//				// No more tokens to transfer.
//				close(ch.sem)
//				ch.sem = newSem
//				return
//			}
//		}
//	}
func (ch *ConcurrencyHandler) ResizeSemaphore(newSize int64) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	if newSize == ch.currentCapacity {
		return // No change needed
	}

	// Ensure the active requests are below the new capacity before resizing
	ch.waitForCapacityAdjustment(newSize)

	// Create a new semaphore with the new size
	newSem := semaphore.NewWeighted(newSize)

	// Migrate the current permits to the new semaphore
	for i := int64(0); i < ch.activePermits && i < newSize; i++ {
		newSem.Release(1)
	}

	// Replace the old semaphore with the new one
	ch.sem = newSem
	ch.currentCapacity = newSize
	ch.logger.Info("Semaphore resized", zap.Int64("newCapacity", newSize))
}

// waitForCapacityAdjustment blocks until the number of active permits is reduced to the new capacity.
func (ch *ConcurrencyHandler) waitForCapacityAdjustment(newSize int64) {
	for {
		ch.lock.RLock()
		if ch.activePermits <= newSize {
			ch.lock.RUnlock()
			break
		}
		ch.lock.RUnlock()
		time.Sleep(100 * time.Millisecond) // Sleep to throttle the loop and reduce CPU usage
		ch.logger.Info("Waiting for active requests to reduce", zap.Int64("currentActive", ch.activePermits), zap.Int64("targetCapacity", newSize))
	}
}
