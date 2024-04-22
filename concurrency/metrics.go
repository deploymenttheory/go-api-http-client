// concurrency/metrics.go
package concurrency

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// EvaluateAndAdjustConcurrency evaluates the HTTP response from a server along with the request's response time
// and adjusts the concurrency level of the system accordingly. It utilizes three monitoring functions:
// MonitorRateLimitHeaders, MonitorServerResponseCodes, and MonitorResponseTimeVariability, each of which
// provides feedback on different aspects of the response and system's current state. The function aggregates
// feedback from these monitoring functions to make a decision on whether to scale up or scale down the concurrency.
// The decision is based on a simple majority of suggestions: if more functions suggest scaling down (return -1),
// it scales down; if more suggest scaling up (return 1), it scales up. This method centralizes concurrency control
// decision-making, providing a systematic approach to managing request handling capacity based on real-time
// operational metrics.
//
// Parameters:
//
//	resp - The HTTP response received from the server.
//	responseTime - The time duration between sending the request and receiving the response.
//
// It logs the specific reason for scaling decisions, helping in traceability and fine-tuning system performance.
func (ch *ConcurrencyHandler) EvaluateAndAdjustConcurrency(resp *http.Response, responseTime time.Duration) {
	rateLimitFeedback := ch.MonitorRateLimitHeaders(resp)
	responseCodeFeedback := ch.MonitorServerResponseCodes(resp)
	responseTimeFeedback := ch.MonitorResponseTimeVariability(responseTime)

	// Compute the weighted feedback
	weightedFeedback := float64(rateLimitFeedback)*WeightRateLimit +
		float64(responseCodeFeedback)*WeightResponseCodes +
		float64(responseTimeFeedback)*WeightResponseTime

	// Log the feedback and weighted result for debugging
	ch.logger.Debug("Concurrency Adjustment Feedback",
		zap.Float64("WeightedFeedback", weightedFeedback))

	// Apply thresholds to determine scaling action
	if weightedFeedback <= ThresholdScaleDown {
		ch.logger.Info("Scaling down the concurrency", zap.Float64("WeightedFeedback", weightedFeedback))
		ch.ScaleDown()
	} else if weightedFeedback >= ThresholdScaleUp {
		ch.logger.Info("Scaling up the concurrency", zap.Float64("WeightedFeedback", weightedFeedback))
		ch.ScaleUp()
	} else {
		ch.logger.Info("Maintaining current concurrency level", zap.Float64("WeightedFeedback", weightedFeedback))
	}
}

// MonitorRateLimitHeaders monitors the rate limit headers in the response and suggests a concurrency adjustment.
// func (ch *ConcurrencyHandler) MonitorRateLimitHeaders(resp *http.Response) int {
// 	remaining := resp.Header.Get("X-RateLimit-Remaining")
// 	retryAfter := resp.Header.Get("Retry-After")
// 	suggestion := 0

// 	if remaining != "" {
// 		remainingValue, err := strconv.Atoi(remaining)
// 		if err == nil && remainingValue < 10 {
// 			// Suggest decrease concurrency if X-RateLimit-Remaining is below the threshold
// 			suggestion = -1
// 		}
// 	}

// 	if retryAfter != "" {
// 		// Suggest decrease concurrency if Retry-After is specified
// 		suggestion = -1
// 	} else {
// 		// Suggest increase concurrency if currently below maximum limit and no other decrease suggestion has been made
// 		if len(ch.sem) < MaxConcurrency && suggestion == 0 {
// 			suggestion = 1
// 		}
// 	}

//		return suggestion
//	}
func (ch *ConcurrencyHandler) MonitorRateLimitHeaders(resp *http.Response) int {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	retryAfter := resp.Header.Get("Retry-After")
	suggestion := 0

	if remaining != "" {
		remainingValue, err := strconv.Atoi(remaining)
		if err == nil && remainingValue < 10 {
			suggestion = -1
		}
	}

	if retryAfter != "" {
		suggestion = -1
	}

	return suggestion
}

// MonitorServerResponseCodes monitors the response status codes and suggests a concurrency adjustment.
// func (ch *ConcurrencyHandler) MonitorServerResponseCodes(resp *http.Response) int {
// 	statusCode := resp.StatusCode

// 	// Lock the metrics to ensure thread safety
// 	ch.Metrics.Lock.Lock()
// 	defer ch.Metrics.Lock.Unlock()

// 	// Update the appropriate error count based on the response status code
// 	switch {
// 	case statusCode >= 500 && statusCode < 600:
// 		ch.Metrics.TotalRateLimitErrors++
// 	case statusCode >= 400 && statusCode < 500:
// 		// Assuming 4xx errors as client errors
// 		ch.Metrics.TotalRetries++
// 	}

// 	// Calculate error rate
// 	totalRequests := float64(ch.Metrics.TotalRequests)
// 	totalErrors := float64(ch.Metrics.TotalRateLimitErrors + ch.Metrics.TotalRetries)
// 	errorRate := totalErrors / totalRequests

// 	// Set the new error rate in the metrics
// 	ch.Metrics.ResponseCodeMetrics.ErrorRate = errorRate

//		// Determine action based on the error rate
//		if errorRate > ErrorRateThreshold {
//			// Suggest decrease concurrency
//			return -1
//		} else if errorRate <= ErrorRateThreshold && len(ch.sem) < MaxConcurrency {
//			// Suggest increase concurrency if there is capacity
//			return 1
//		}
//		return 0
//	}
func (ch *ConcurrencyHandler) MonitorServerResponseCodes(resp *http.Response) int {
	statusCode := resp.StatusCode
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	if statusCode >= 500 {
		ch.Metrics.TotalRateLimitErrors++
		return -1
	} else if statusCode >= 400 {
		ch.Metrics.TotalRetries++
		return -1
	}

	return 0
}

// MonitorResponseTimeVariability monitors the response time variability and suggests a concurrency adjustment.
// func (ch *ConcurrencyHandler) MonitorResponseTimeVariability(responseTime time.Duration) int {
// 	ch.Metrics.Lock.Lock()
// 	defer ch.Metrics.Lock.Unlock()

// 	// Update ResponseTimeVariability metrics
// 	ch.Metrics.ResponseTimeVariability.Lock.Lock()
// 	defer ch.Metrics.ResponseTimeVariability.Lock.Unlock()
// 	ch.Metrics.ResponseTimeVariability.Total += responseTime
// 	ch.Metrics.ResponseTimeVariability.Count++

// 	// Calculate average response time
// 	ch.Metrics.ResponseTimeVariability.Average = ch.Metrics.ResponseTimeVariability.Total / time.Duration(ch.Metrics.ResponseTimeVariability.Count)

// 	// Calculate variance of response times
// 	ch.Metrics.ResponseTimeVariability.Variance = ch.calculateVariance(ch.Metrics.ResponseTimeVariability.Average, responseTime)

// 	// Calculate standard deviation of response times
// 	stdDev := math.Sqrt(ch.Metrics.ResponseTimeVariability.Variance)

//		// Determine action based on standard deviation
//		if stdDev > ch.Metrics.ResponseTimeVariability.StdDevThreshold {
//			// Suggest decrease concurrency
//			return -1
//		} else if stdDev <= ch.Metrics.ResponseTimeVariability.StdDevThreshold && len(ch.sem) < MaxConcurrency {
//			// Suggest increase concurrency if there is capacity
//			return 1
//		}
//		return 0
//	}
func (ch *ConcurrencyHandler) MonitorResponseTimeVariability(responseTime time.Duration) int {
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Update total response time and count
	ch.Metrics.ResponseTimeVariability.Total += responseTime
	ch.Metrics.ResponseTimeVariability.Count++

	// Calculate the average response time
	averageResponseTime := ch.Metrics.ResponseTimeVariability.Total / time.Duration(ch.Metrics.ResponseTimeVariability.Count)

	// Calculate variance
	variance := ch.calculateVariance(averageResponseTime, responseTime)
	// Calculate standard deviation
	stdDev := math.Sqrt(variance)

	// Convert MaxAcceptableResponseTimeVariability to seconds for comparison
	maxStdDev := MaxAcceptableResponseTimeVariability.Seconds()

	if stdDev > maxStdDev {
		return -1 // Suggest to decrease concurrency if stdDev exceeds the maximum threshold
	}
	return 1 // Suggest to increase concurrency if stdDev is within the acceptable range
}

// calculateVariance calculates the variance between the average response time and a new sample.
func (ch *ConcurrencyHandler) calculateVariance(average, newSample time.Duration) float64 {
	mean := average.Seconds()       // Convert to seconds
	newValue := newSample.Seconds() // Convert to seconds
	newVariance := (float64(ch.Metrics.ResponseTimeVariability.Count-1)*math.Pow(mean-newValue, 2) + ch.Metrics.ResponseTimeVariability.Variance) / float64(ch.Metrics.ResponseTimeVariability.Count)
	ch.Metrics.ResponseTimeVariability.Variance = newVariance // Update the variance in metrics
	return newVariance
}
