// concurrencyhandler/metrics.go
package concurrencyhandler

import "time"

// AverageAcquisitionTime computes the average time taken to acquire a token
// from the semaphore. It helps in understanding the contention for tokens
// and can be used to adjust concurrency limits.
func (ch *ConcurrencyHandler) AverageAcquisitionTime() time.Duration {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	if len(ch.AcquisitionTimes) == 0 {
		return 0
	}

	totalTime := time.Duration(0)
	for _, t := range ch.AcquisitionTimes {
		totalTime += t
	}
	return totalTime / time.Duration(len(ch.AcquisitionTimes))
}

// HistoricalAverageAcquisitionTime computes the average time taken to acquire
// a token from the semaphore over a historical period (e.g., the last 5 minutes).
// It helps in understanding the historical contention for tokens and can be used
// to adjust concurrency limits.
func (ch *ConcurrencyHandler) HistoricalAverageAcquisitionTime() time.Duration {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// For simplicity, let's say we store the last 5 minutes of acquisition times.
	// This means if EvaluationInterval is 1 minute, we consider the last 5 data points.
	historicalCount := 5
	if len(ch.AcquisitionTimes) < historicalCount {
		return ch.AverageAcquisitionTime() // If not enough historical data, return the overall average
	}

	totalTime := time.Duration(0)
	for _, t := range ch.AcquisitionTimes[len(ch.AcquisitionTimes)-historicalCount:] {
		totalTime += t
	}
	return totalTime / time.Duration(historicalCount)
}

// updatePerformanceMetrics updates the ConcurrencyHandler's performance metrics
// by recording the duration of an HTTP request and incrementing the total
// request count. This function is thread-safe and uses a mutex to synchronize
// updates to the performance metrics.
//
// Parameters:
// - duration: The time duration it took for an HTTP request to complete.
//
// This function should be called after each HTTP request to keep track of the
// ConcurrencyHandler's performance over time.
func (ch *ConcurrencyHandler) UpdatePerformanceMetrics(duration time.Duration) {
	ch.PerfMetrics.lock.Lock()
	defer ch.PerfMetrics.lock.Unlock()
	ch.PerfMetrics.TotalResponseTime += duration
	ch.PerfMetrics.TotalRequests++
}

// Min returns the smaller of the two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
