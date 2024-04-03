// concurrencyhandler/dynamic_token_adjustment.go

package concurrencyhandler

import (
	"time"

	"go.uber.org/zap"
)

// AdjustConcurrencyLimit dynamically modifies the maximum concurrency limit
// based on the newLimit provided. This function helps in adjusting the concurrency
// limit in real-time based on observed system performance and other metrics. It
// transfers the tokens from the old semaphore to the new one, ensuring that there's
// no loss of tokens during the transition.
func (ch *ConcurrencyHandler) AdjustConcurrencyLimit(newLimit int) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	if newLimit <= 0 {
		return // Avoid setting a non-positive limit
	}

	// Create a new semaphore with the desired limit
	newSem := make(chan struct{}, newLimit)

	// Transfer tokens from the old semaphore to the new one
	for i := 0; i < len(ch.sem) && i < newLimit; i++ {
		newSem <- struct{}{}
	}

	ch.sem = newSem
}

// AdjustConcurrencyBasedOnMetrics evaluates the current metrics and adjusts the
// concurrency limit if required. It checks metrics like average token acquisition
// time and decides on a new concurrency limit. The method ensures that the new
// limit respects the minimum and maximum allowed concurrency bounds.
func (ch *ConcurrencyHandler) AdjustConcurrencyBasedOnMetrics() {
	// Calculate the average acquisition time
	avgAcquisitionTime := ch.AverageAcquisitionTime()

	// Get the current concurrency limit
	currentLimit := cap(ch.sem)

	// Calculate the historical average acquisition time
	historicalAvgAcquisitionTime := ch.HistoricalAverageAcquisitionTime()

	// Decide on a new limit based on metrics
	newLimit := currentLimit
	if avgAcquisitionTime > time.Duration(float64(historicalAvgAcquisitionTime)*1.2) { // 20% increase in acquisition time
		newLimit = currentLimit - 2 // decrease concurrency more aggressively
	} else if avgAcquisitionTime < time.Duration(float64(historicalAvgAcquisitionTime)*0.8) { // 20% decrease in acquisition time
		newLimit = currentLimit + 2 // increase concurrency more aggressively
	} else if avgAcquisitionTime > historicalAvgAcquisitionTime {
		newLimit = currentLimit - 1 // decrease concurrency conservatively
	} else if avgAcquisitionTime < historicalAvgAcquisitionTime {
		newLimit = currentLimit + 1 // increase concurrency conservatively
	}

	// Ensure newLimit is within safety bounds
	if newLimit > MaxConcurrency {
		newLimit = MaxConcurrency
	} else if newLimit < MinConcurrency {
		newLimit = MinConcurrency
	}

	// Adjust concurrency if the new limit is different from the current
	if newLimit != currentLimit {
		ch.AdjustConcurrencyLimit(newLimit)

		// Log the adjustment
		ch.logger.Debug("Adjusted concurrency",
			zap.Int("OldLimit", currentLimit),
			zap.Int("NewLimit", newLimit),
			zap.String("Reason", "Based on average acquisition time"),
			zap.Duration("AverageAcquisitionTime", avgAcquisitionTime),
			zap.Duration("HistoricalAverageAcquisitionTime", historicalAvgAcquisitionTime),
		)
	}
}

// EvaluateMetricsAndAdjustConcurrency evaluates the performance metrics and makes necessary
// adjustments to the concurrency limit. The method assesses the average response time
// and adjusts the concurrency based on how it compares to the historical average acquisition time.
// If the average response time has significantly increased compared to the historical average,
// the concurrency limit is decreased, and vice versa. The method ensures that the concurrency
// limit remains within the bounds defined by the system's best practices.
func (ch *ConcurrencyHandler) EvaluateMetricsAndAdjustConcurrency() {
	ch.PerfMetrics.lock.Lock()
	averageResponseTime := ch.PerfMetrics.TotalResponseTime / time.Duration(ch.PerfMetrics.TotalRequests)
	ch.PerfMetrics.lock.Unlock()

	historicalAverageAcquisitionTime := ch.HistoricalAverageAcquisitionTime()

	// Decide on the new limit based on the average response time compared to the historical average
	newLimit := cap(ch.sem) // Start with the current limit
	if averageResponseTime > time.Duration(float64(historicalAverageAcquisitionTime)*1.2) {
		// Decrease concurrency more aggressively if the average response time has significantly increased
		newLimit -= 1
	} else if averageResponseTime < time.Duration(float64(historicalAverageAcquisitionTime)*0.8) {
		// Increase concurrency more aggressively if the average response time has significantly decreased
		newLimit += 1
	}

	// Ensure the new limit is within the defined bounds
	if newLimit > MaxConcurrency {
		newLimit = MaxConcurrency
	} else if newLimit < MinConcurrency {
		newLimit = MinConcurrency
	}

	// Adjust the concurrency limit if the new limit is different from the current limit
	currentLimit := cap(ch.sem)
	if newLimit != currentLimit {
		ch.AdjustConcurrencyLimit(newLimit)

		// Log the adjustment for debugging purposes
		ch.logger.Debug("Adjusted concurrency",
			zap.Int("OldLimit", currentLimit),
			zap.Int("NewLimit", newLimit),
			zap.String("Reason", "Based on evaluation of metrics"),
			zap.Duration("AverageResponseTime", averageResponseTime),
			zap.Duration("HistoricalAverageAcquisitionTime", historicalAverageAcquisitionTime),
		)
	}
}

//------ Concurrency Monitoring Functions:

// StartMetricEvaluation continuously monitors the client's interactions with the API and adjusts the concurrency limits dynamically.
// The function evaluates metrics at regular intervals to detect burst activity patterns.
// If a burst activity is detected (e.g., many requests in a short period), the evaluation interval is reduced for more frequent checks.
// Otherwise, it reverts to a default interval for regular checks.
// After each evaluation, the function calls EvaluateMetricsAndAdjustConcurrency to potentially adjust the concurrency based on observed metrics.
//
// The evaluation process works as follows:
// 1. Sleep for the defined evaluation interval.
// 2. Check if there's a burst in activity using the isBurstActivity method.
// 3. If a burst is detected, the evaluation interval is shortened to more frequently monitor and adjust the concurrency.
// 4. If no burst is detected, it maintains the default evaluation interval.
// 5. It then evaluates the metrics and adjusts the concurrency accordingly.
func (ch *ConcurrencyHandler) StartMetricEvaluation() {
	evalInterval := 5 * time.Minute // Initial interval

	for {
		time.Sleep(evalInterval) // Wait for the defined evaluation interval

		// Determine if there's been a burst in activity
		if ch.isBurstActivity() {
			evalInterval = 1 * time.Minute // Increase the frequency of checks during burst activity
		} else {
			evalInterval = 5 * time.Minute // Revert to the default interval outside of burst periods
		}

		// Evaluate the current metrics and adjust concurrency if necessary
		ch.EvaluateMetricsAndAdjustConcurrency()
	}
}

// isBurstActivity checks if there's been a burst in activity based on the acquisition times of the tokens.
// A burst is considered to have occurred if the time since the last token acquisition is short.
func (ch *ConcurrencyHandler) isBurstActivity() bool {
	// Lock before checking the last token acquisition time
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// Consider it a burst if the last token was acquired less than 2 minutes ago
	return time.Since(ch.lastTokenAcquisitionTime) < 2*time.Minute
}

// StartConcurrencyAdjustment launches a periodic checker that evaluates current metrics and adjusts concurrency limits if needed.
// It uses a ticker to periodically trigger the adjustment logich.
func (ch *ConcurrencyHandler) StartConcurrencyAdjustment() {
	ticker := time.NewTicker(EvaluationInterval)
	defer ticker.Stop()

	for range ticker.C {
		ch.AdjustConcurrencyBasedOnMetrics()
	}
}

// Returns the average Acquisition Time to get a token from the semaphore
func (ch *ConcurrencyHandler) GetAverageAcquisitionTime() time.Duration {
	// Assuming ConcurrencyMgr has a method to get this metric
	return ch.AverageAcquisitionTime()
}

func (ch *ConcurrencyHandler) GetHistoricalAverageAcquisitionTime() time.Duration {
	// Assuming ConcurrencyMgr has a method to get this metric
	return ch.HistoricalAverageAcquisitionTime()
}

// GetPerformanceMetrics returns the current performance metrics of the ConcurrencyHandler.
// This includes counts of total requests, retries, rate limit errors, total response time,
// and token wait time.
func (ch *ConcurrencyHandler) GetPerformanceMetrics() *PerformanceMetrics {
	return ch.PerfMetrics
}
