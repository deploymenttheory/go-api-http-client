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
	// Lock to ensure thread safety
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// We must consider the current capacity rather than the length of the semaphore
	currentSize := ch.currentCapacity
	if currentSize > MinConcurrency {
		newSize := currentSize - 1
		ch.logger.Info("Reducing request concurrency", zap.Int64("currentSize", currentSize), zap.Int64("newSize", newSize))
		ch.ResizeSemaphore(newSize)
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
		// Scale up by 10% of the available margin, ensuring we do not exceed MaxConcurrency
		newSize := currentSize + int64(float64(MaxConcurrency-currentSize)*0.1)
		newSize = min(newSize, MaxConcurrency)
		ch.logger.Info("Increasing request concurrency", zap.Int64("currentSize", currentSize), zap.Int64("newSize", newSize))
		ch.ResizeSemaphore(newSize)
	} else {
		ch.logger.Info("Concurrency already at maximum level; cannot increase further", zap.Int64("currentSize", currentSize))
	}
}
