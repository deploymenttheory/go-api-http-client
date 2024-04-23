// concurrency/metrics.go
package concurrency

import (
	"math"
	"net/http"
	"strconv"
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

	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

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

// MonitorResponseTimeVariability monitors the response time variability and suggests a concurrency adjustment.
func (ch *ConcurrencyHandler) MonitorResponseTimeVariability(responseTime time.Duration) int {
	// Append the latest response time
	responseTimes = append(responseTimes, responseTime)
	if len(responseTimes) > 5 { // Use the last 5 measurements
		responseTimes = responseTimes[1:]
	}

	// Calculate average response time from the slice
	var sum time.Duration
	for _, rt := range responseTimes {
		sum += rt
	}
	averageResponseTime := sum / time.Duration(len(responseTimes))

	// Calculate standard deviation based on the moving average
	var varianceSum float64
	for _, rt := range responseTimes {
		varianceSum += math.Pow(rt.Seconds()-averageResponseTime.Seconds(), 2)
	}
	variance := varianceSum / float64(len(responseTimes))
	stdDev := math.Sqrt(variance)

	// Determine action based on standard deviation against a higher threshold
	if stdDev > (ch.Metrics.ResponseTimeVariability.StdDevThreshold * 1.5) { // Increased threshold
		return -1 // Suggest decrease concurrency
	} else if stdDev <= ch.Metrics.ResponseTimeVariability.StdDevThreshold && len(ch.sem) < MaxConcurrency {
		return 1 // Suggest increase concurrency if there is capacity
	}
	return 0
}

// calculateVariance calculates the variance of response times.
func (ch *ConcurrencyHandler) calculateVariance(averageResponseTime time.Duration, responseTime time.Duration) float64 {
	// Convert time.Duration values to seconds
	averageSeconds := averageResponseTime.Seconds()
	responseSeconds := responseTime.Seconds()

	// Calculate variance
	variance := (float64(ch.Metrics.ResponseTimeVariability.Count-1)*math.Pow(averageSeconds-responseSeconds, 2) + ch.Metrics.ResponseTimeVariability.Variance) / float64(ch.Metrics.ResponseTimeVariability.Count)
	ch.Metrics.ResponseTimeVariability.Variance = variance
	return variance
}
