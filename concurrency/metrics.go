// concurrency/metrics.go
package concurrency

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Defined weights for the metrics
var metricWeights = map[string]float64{
	"RateLimit":    5.0, // High importance
	"ServerError":  3.0, // High importance
	"ResponseTime": 1.0, // Lower importance
}

// EvaluateAndAdjustConcurrency assesses the current state of system metrics and decides whether to scale
// up or down the number of concurrent operations allowed. It employs a combination of strategies:
// a weighted scoring system, threshold-based direct actions, and cumulative impact assessment.
//
// A weighted scoring system is used to prioritize the importance of different system metrics. Each metric
// can influence the scaling decision based on its assigned weight, reflecting its relative impact on system performance.
//
// Threshold-based scaling provides a fast-track decision path for critical metrics that have exceeded predefined limits.
// If a critical metric, such as the rate limit remaining slots or server error rates, crosses a specified threshold,
// immediate action is taken to scale down the concurrency to prevent system overload.
//
// Cumulative impact assessment calculates a cumulative score from all monitored metrics, taking into account
// their respective weights. This score determines the overall tendency of the system to either scale up or down.
// If the score indicates a negative trend (i.e., below zero), the system will scale down to reduce load.
// Conversely, a positive score suggests that there is capacity to handle more concurrent operations, leading
// to a scale-up decision.
//
// Parameters:
//   - resp: The HTTP response received from the server, providing status codes and headers for rate limiting.
//   - responseTime: The time duration between sending the request and receiving the response, indicating the server's responsiveness.
//
// The function logs the decision process at each step, providing traceability and insight into the scaling mechanism.
// The method should be called after each significant interaction with the external system (e.g., an HTTP request) to
// ensure concurrency levels are adapted to current conditions.
//
// Returns: None. The function directly calls the ScaleUp or ScaleDown methods as needed.
//
// Note: This function does not return any value; it performs actions based on internal assessments and logs outcomes.
func (ch *ConcurrencyHandler) EvaluateAndAdjustConcurrency(resp *http.Response, responseTime time.Duration) {
	rateLimitFeedback := ch.MonitorRateLimitHeaders(resp)
	responseCodeFeedback := ch.MonitorServerResponseCodes(resp)
	responseTimeFeedback := ch.MonitorResponseTimeVariability(responseTime)

	// Use weighted scores for each metric.
	weightedRateLimitScore := float64(rateLimitFeedback) * metricWeights["RateLimit"]
	weightedResponseCodeScore := float64(responseCodeFeedback) * metricWeights["ServerError"]
	weightedResponseTimeScore := float64(responseTimeFeedback) * metricWeights["ResponseTime"]

	// Calculate the cumulative score.
	cumulativeScore := weightedRateLimitScore + weightedResponseCodeScore + weightedResponseTimeScore

	// Detailed debugging output
	ch.logger.Debug("Evaluate and Adjust Concurrency",
		zap.String("event", "EvaluateConcurrency"),
		zap.Float64("weightedRateLimitScore", weightedRateLimitScore),
		zap.Float64("weightedResponseCodeScore", weightedResponseCodeScore),
		zap.Float64("weightedResponseTimeScore", weightedResponseTimeScore),
		zap.Float64("cumulativeScore", cumulativeScore),
		zap.Int("rateLimitFeedback", rateLimitFeedback),
		zap.Int("responseCodeFeedback", responseCodeFeedback),
		zap.Int("responseTimeFeedback", responseTimeFeedback),
		zap.Duration("responseTime", responseTime),
	)

	// Check for successful response and log appropriately
	if responseCodeFeedback == 1 { // Assuming 1 indicates success
		ch.logger.Info("Successful response noted, checking for scaling necessity.",
			zap.String("API Response", "Success"),
			zap.Int("StatusCode", resp.StatusCode),
		)
	}

	// Check critical thresholds
	if rateLimitFeedback <= RateLimitCriticalThreshold || responseCodeFeedback < 0 {
		if weightedRateLimitScore >= ErrorResponseThreshold || weightedResponseCodeScore >= ErrorResponseThreshold {
			ch.logger.Warn("Scaling down due to critical threshold breach",
				zap.String("event", "CriticalThresholdBreach"),
				zap.Int("rateLimitFeedback", rateLimitFeedback),
				zap.Float64("errorResponseRate", weightedResponseCodeScore),
			)
			ch.ScaleDown()
			return
		}
	}

	// Evaluate cumulative impact and make a scaling decision based on the cumulative score and other metrics.
	if cumulativeScore < 0 {
		utilizedBefore := len(ch.sem) // Tokens in use before scaling down.
		ch.ScaleDown()
		utilizedAfter := len(ch.sem) // Tokens in use after scaling down.
		ch.logger.Info("Concurrency scaling decision: scale down.",
			zap.Float64("cumulativeScore", cumulativeScore),
			zap.Int("utilizedTokensBefore", utilizedBefore),
			zap.Int("utilizedTokensAfter", utilizedAfter),
			zap.Int("availableTokensBefore", cap(ch.sem)-utilizedBefore),
			zap.Int("availableTokensAfter", cap(ch.sem)-utilizedAfter),
			zap.String("reason", "Cumulative impact of metrics suggested an overload."),
		)
	} else if cumulativeScore > 0 {
		utilizedBefore := len(ch.sem) // Tokens in use before scaling up.
		ch.ScaleUp()
		utilizedAfter := len(ch.sem) // Tokens in use after scaling up.
		ch.logger.Info("Concurrency scaling decision: scale up.",
			zap.Float64("cumulativeScore", cumulativeScore),
			zap.Int("utilizedTokensBefore", utilizedBefore),
			zap.Int("utilizedTokensAfter", utilizedAfter),
			zap.Int("availableTokensBefore", cap(ch.sem)-utilizedBefore),
			zap.Int("availableTokensAfter", cap(ch.sem)-utilizedAfter),
			zap.String("reason", "Metrics indicate available resources to handle more load."),
		)
	} else {
		ch.logger.Info("Concurrency scaling decision: no change.",
			zap.Float64("cumulativeScore", cumulativeScore),
			zap.Int("currentUtilizedTokens", len(ch.sem)),
			zap.Int("currentAvailableTokens", cap(ch.sem)-len(ch.sem)),
			zap.String("reason", "Metrics are stable, maintaining current concurrency level."),
		)
	}
}

// MonitorRateLimitHeaders monitors the rate limit headers in the response and suggests a concurrency adjustment.
func (ch *ConcurrencyHandler) MonitorRateLimitHeaders(resp *http.Response) int {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	retryAfter := resp.Header.Get("Retry-After")
	if remaining == "" && retryAfter == "" {
		// No rate limit information available, return a neutral score
		return 0
	}

	suggestion := 0
	if remaining != "" {
		remainingValue, err := strconv.Atoi(remaining)
		if err == nil && remainingValue < 10 {
			suggestion = -1 // Suggest decrease concurrency if critically low
		}
	}

	if retryAfter != "" {
		suggestion = -1 // Suggest decrease concurrency if Retry-After is specified
	}

	return suggestion
}

// MonitorServerResponseCodes monitors the response status codes and suggests a concurrency adjustment.
func (ch *ConcurrencyHandler) MonitorServerResponseCodes(resp *http.Response) int {
	statusCode := resp.StatusCode

	ch.Metrics.Lock()
	defer ch.Metrics.Unlock()

	if statusCode >= 200 && statusCode < 300 {
		// Reset error counts for successful responses
		ch.Metrics.TotalRateLimitErrors = 0
		ch.Metrics.TotalRetries = 0
		return 0 // No need to adjust concurrency for successful responses
	} else if statusCode >= 500 && statusCode < 600 {
		ch.Metrics.TotalRateLimitErrors++
	} else if statusCode >= 400 && statusCode < 500 {
		ch.Metrics.TotalRetries++
	}

	totalRequests := float64(ch.Metrics.TotalRequests)
	totalErrors := float64(ch.Metrics.TotalRateLimitErrors + ch.Metrics.TotalRetries)
	if totalErrors == 0 {
		return 0 // No errors, no concurrency adjustment needed
	}

	errorRate := totalErrors / totalRequests
	ch.Metrics.ResponseCodeMetrics.ErrorRate = errorRate

	ch.logger.Debug("Server Response Code Monitoring",
		zap.Int("StatusCode", statusCode),
		zap.Float64("TotalRequests", totalRequests),
		zap.Float64("TotalErrors", totalErrors),
		zap.Float64("ErrorRate", errorRate),
	)

	// Only suggest a scale-down if the error rate exceeds the threshold
	if errorRate > ErrorRateThreshold {
		return -1 // Suggest decrease concurrency
	}
	return 0 // Default to no change if error rate is within acceptable limits
}

// A slice to hold the last n response times for averaging
var responseTimes []time.Duration
var responseTimesLock sync.Mutex

// MonitorResponseTimeVariability assesses the response time variability from a series of HTTP requests and decides whether to adjust the concurrency level of outgoing requests. This function is integral to maintaining optimal system performance under varying load conditions.
//
// The function first appends the latest response time to a sliding window of the last 10 response times to maintain a recent history. It then calculates the standard deviation and the average of these times. The standard deviation helps determine the variability or consistency of response times, while the average gives a central tendency.
//
// Based on these calculated metrics, the function employs a multi-factor decision mechanism:
// - If the standard deviation exceeds a pre-defined threshold and the average response time is greater than an acceptable maximum, a debounce counter is incremented. This counter must reach a predefined threshold (debounceScaleDownThreshold) before a decision to decrease concurrency is made, ensuring that only sustained negative trends lead to a scale down.
// - If the standard deviation is below or equal to the threshold, suggesting stable response times, and the system is currently operating below its concurrency capacity, it may suggest an increase in concurrency to improve throughput.
//
// This approach aims to prevent transient spikes in response times from causing undue scaling actions, thus stabilizing the overall performance and responsiveness of the system.
//
// Returns:
// - (-1) to suggest a decrease in concurrency,
// - (1) to suggest an increase in concurrency,
// - (0) to indicate no change needed.
func (ch *ConcurrencyHandler) MonitorResponseTimeVariability(responseTime time.Duration) int {
	ch.Metrics.ResponseTimeVariability.Lock()
	defer ch.Metrics.ResponseTimeVariability.Unlock()

	responseTimesLock.Lock() // Ensure safe concurrent access
	responseTimes = append(responseTimes, responseTime)
	if len(responseTimes) > 10 {
		responseTimes = responseTimes[1:] // Maintain last 10 measurements
	}
	responseTimesLock.Unlock()

	stdDev := calculateStdDev(responseTimes)
	averageResponseTime := calculateAverage(responseTimes)

	// Check if conditions suggest a need to scale down
	if stdDev > ch.Metrics.ResponseTimeVariability.StdDevThreshold && averageResponseTime > AcceptableAverageResponseTime {
		ch.Metrics.ResponseTimeVariability.DebounceScaleDownCount++
		if ch.Metrics.ResponseTimeVariability.DebounceScaleDownCount >= debounceScaleDownThreshold {
			ch.Metrics.ResponseTimeVariability.DebounceScaleDownCount = 0
			return -1 // Suggest decrease concurrency
		}
	} else {
		ch.Metrics.ResponseTimeVariability.DebounceScaleDownCount = 0 // Reset counter if conditions are not met
	}

	// Check if conditions suggest a need to scale up
	if stdDev <= ch.Metrics.ResponseTimeVariability.StdDevThreshold && averageResponseTime <= AcceptableAverageResponseTime {
		ch.Metrics.ResponseTimeVariability.DebounceScaleUpCount++
		if ch.Metrics.ResponseTimeVariability.DebounceScaleUpCount >= debounceScaleDownThreshold {
			ch.Metrics.ResponseTimeVariability.DebounceScaleUpCount = 0
			return 1 // Suggest increase concurrency
		}
	} else {
		ch.Metrics.ResponseTimeVariability.DebounceScaleUpCount = 0 // Reset counter if conditions are not met
	}

	return 0 // Default to no change
}

// calculateAverage computes the average response time from a slice of time.Duration values.
// The average, or mean, is a measure of the central tendency of a set of values, providing a simple
// summary of the 'typical' value in a set. In the context of response times, the average gives a
// straightforward indication of the overall system response performance over a given set of requests.
//
// The function performs the following steps to calculate the average response time:
// 1. Sum all the response times in the input slice.
// 2. Divide the total sum by the number of response times to find the mean value.
//
// This method of averaging is vital for assessing the overall health and efficiency of the system under load.
// Monitoring average response times can help in identifying trends in system performance, guiding capacity planning,
// and optimizing resource allocation.
//
// Parameters:
// - times: A slice of time.Duration values, each representing the response time for a single request.
//
// Returns:
//   - time.Duration: The average response time across all provided times. This is a time.Duration value that
//     can be directly compared to other durations or used to set thresholds for alerts or further analysis.
//
// Example Usage:
// This function is typically used in performance analysis where the average response time is monitored over
// specific intervals to ensure service level agreements (SLAs) are met or to trigger scaling actions when
// average response times exceed acceptable levels.
func calculateAverage(times []time.Duration) time.Duration {
	var total time.Duration
	for _, t := range times {
		total += t
	}
	return total / time.Duration(len(times))
}

// calculateStdDev computes the standard deviation of response times from a slice of time.Duration values.
// Standard deviation is a measure of the amount of variation or dispersion in a set of values. A low standard
// deviation indicates that the values tend to be close to the mean (average) of the set, while a high standard
// deviation indicates that the values are spread out over a wider range.
//
// The function performs the following steps to calculate the standard deviation:
// 1. Calculate the mean (average) response time of the input slice.
// 2. Sum the squared differences from the mean for each response time. This measures the total variance from the mean.
// 3. Divide the total variance by the number of response times to get the average variance.
// 4. Take the square root of the average variance to obtain the standard deviation.
//
// This statistical approach is crucial for identifying how consistently the system responds under different loads
// and can be instrumental in diagnosing performance fluctuations in real-time systems.
//
// Parameters:
// - times: A slice of time.Duration values representing response times.
//
// Returns:
// - float64: The calculated standard deviation of the response times, which represents the variability in response times.
//
// This function is typically used in performance monitoring to adjust system concurrency based on the stability
// of response times, as part of a larger strategy to optimize application responsiveness and reliability.
func calculateStdDev(times []time.Duration) float64 {
	var sum time.Duration
	for _, t := range times {
		sum += t
	}
	avg := sum / time.Duration(len(times))

	var varianceSum float64
	for _, t := range times {
		varianceSum += math.Pow(t.Seconds()-avg.Seconds(), 2)
	}
	variance := varianceSum / float64(len(times))
	stdDev := math.Sqrt(variance)

	return stdDev
}
