// concurrency/scale.go
package concurrency

import "go.uber.org/zap"

// ScaleDown reduces the concurrency level by one, down to the minimum limit.
func (ch *ConcurrencyHandler) ScaleDown() {
	ch.Lock()
	defer ch.Unlock()

	currentSize := cap(ch.sem)
	if currentSize > MinConcurrency {
		newSize := currentSize - 1
		ch.logger.Info("Reducing request concurrency", zap.Int("currentSize", currentSize), zap.Int("newSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else {
		ch.logger.Info("Concurrency already at minimum level; cannot reduce further", zap.Int("currentSize", currentSize))
	}
}

// ScaleUp increases the concurrency level by one, up to the maximum limit.
func (ch *ConcurrencyHandler) ScaleUp() {
	ch.Lock()
	defer ch.Unlock()

	currentSize := cap(ch.sem)
	if currentSize < MaxConcurrency {
		newSize := currentSize + 1
		ch.logger.Info("Increasing request concurrency", zap.Int("currentSize", currentSize), zap.Int("newSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else {
		ch.logger.Info("Concurrency already at maximum level; cannot increase further", zap.Int("currentSize", currentSize))
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

	for {
		select {
		case token := <-ch.sem:
			select {
			case newSem <- token:
			default:
				ch.sem <- token
			}
		default:
			close(ch.sem)
			ch.sem = newSem
			return
		}
	}
}
