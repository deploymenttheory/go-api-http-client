// concurrency/scale.go
package concurrency

import "go.uber.org/zap"

// ScaleDown reduces the concurrency level by one, down to the minimum limit.
// func (ch *ConcurrencyHandler) ScaleDown() {
// 	// Lock to ensure thread safety
// 	ch.lock.Lock()
// 	defer ch.lock.Unlock()

//		// We must consider the capacity rather than the length of the semaphore channel
//		currentSize := cap(ch.sem)
//		if currentSize > MinConcurrency {
//			newSize := currentSize - 1
//			ch.logger.Info("Reducing request concurrency", zap.Int("currentSize", currentSize), zap.Int("newSize", newSize))
//			ch.ResizeSemaphore(newSize)
//		} else {
//			ch.logger.Info("Concurrency already at minimum level; cannot reduce further", zap.Int("currentSize", currentSize))
//		}
//	}
func (ch *ConcurrencyHandler) ScaleDown() {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	currentSize := ch.currentCapacity
	if currentSize > MinConcurrency {
		// Check if active permits allow for scaling down
		if ch.activePermits < currentSize {
			newSize := currentSize - 1
			ch.logger.Info("Reducing request concurrency", zap.Int64("currentSize", currentSize), zap.Int64("newSize", newSize))
			ch.ResizeSemaphore(newSize)
		} else {
			ch.logger.Info("Cannot scale down due to high number of active permits", zap.Int64("currentSize", currentSize), zap.Int64("activePermits", ch.activePermits))
		}
	} else {
		ch.logger.Info("Concurrency already at minimum level; cannot reduce further", zap.Int64("currentSize", currentSize))
	}
}

// ScaleUp increases the concurrency level by one, up to the maximum limit.
// func (ch *ConcurrencyHandler) ScaleUp() {
// 	// Lock to ensure thread safety
// 	ch.lock.Lock()
// 	defer ch.lock.Unlock()

//		currentSize := cap(ch.sem)
//		if currentSize < MaxConcurrency {
//			newSize := currentSize + 1
//			ch.logger.Info("Increasing request concurrency", zap.Int("currentSize", currentSize), zap.Int("newSize", newSize))
//			ch.ResizeSemaphore(newSize)
//		} else {
//			ch.logger.Info("Concurrency already at maximum level; cannot increase further", zap.Int("currentSize", currentSize))
//		}
//	}
func (ch *ConcurrencyHandler) ScaleUp() {
	// Lock to ensure thread safety
	ch.lock.Lock()
	defer ch.lock.Unlock()

	currentSize := ch.currentCapacity
	if currentSize < MaxConcurrency {
		// Calculate the increase based on a percentage of the available margin
		increase := int64(float64(MaxConcurrency-currentSize) * 0.1)
		if increase < 1 {
			increase = 1 // Ensure at least a minimum increase of 1
		}
		newSize := currentSize + increase
		newSize = min(newSize, MaxConcurrency) // Ensure not exceeding max limit
		if newSize > currentSize {             // Check if there is an actual increase
			ch.logger.Info("Increasing request concurrency", zap.Int64("currentSize", currentSize), zap.Int64("newSize", newSize))
			ch.ResizeSemaphore(newSize)
		} else {
			ch.logger.Info("Attempted to increase concurrency but already at or near maximum limit", zap.Int64("currentSize", currentSize), zap.Int64("newSize", newSize))
		}
	} else {
		ch.logger.Info("Concurrency already at maximum level; cannot increase further", zap.Int64("currentSize", currentSize))
	}
}
