// concurrency/scale.go
package concurrency

import "go.uber.org/zap"

// ScaleDown reduces the concurrency level by one, down to the minimum limit.
func (ch *ConcurrencyHandler) ScaleDown() {
	// Lock to ensure thread safety
	ch.Lock()
	defer ch.Unlock()

	// We must consider the capacity rather than the length of the semaphore channel
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
	// Lock to ensure thread safety
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
