// concurrency/resize.go
package concurrency

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
