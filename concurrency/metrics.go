package concurrency

import (
	"math"
	"net/http"
	"strconv"
	"time"
)

// EvaluateAndAdjustConcurrency evaluates the response from the server and adjusts the concurrency level accordingly.
func (ch *ConcurrencyHandler) EvaluateAndAdjustConcurrency(resp *http.Response, responseTime time.Duration) {
	// Call monitoring functions
	rateLimitFeedback := ch.MonitorRateLimitHeaders(resp)
	responseCodeFeedback := ch.MonitorServerResponseCodes(resp)
	responseTimeFeedback := ch.MonitorResponseTimeVariability(responseTime)

	// Determine overall action based on feedback
	suggestions := []int{rateLimitFeedback, responseCodeFeedback, responseTimeFeedback}
	scaleDownCount := 0
	scaleUpCount := 0

	for _, suggestion := range suggestions {
		switch suggestion {
		case -1:
			scaleDownCount++
		case 1:
			scaleUpCount++
		}
	}

	// Decide on scaling action
	if scaleDownCount > scaleUpCount {
		ch.ScaleDown()
	} else if scaleUpCount > scaleDownCount {
		ch.ScaleUp()
	}
}

// MonitorRateLimitHeaders monitors the rate limit headers in the response and suggests a concurrency adjustment.
func (ch *ConcurrencyHandler) MonitorRateLimitHeaders(resp *http.Response) int {
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	retryAfter := resp.Header.Get("Retry-After")
	suggestion := 0

	if remaining != "" {
		remainingValue, err := strconv.Atoi(remaining)
		if err == nil && remainingValue < 10 {
			// Suggest decrease concurrency if X-RateLimit-Remaining is below the threshold
			suggestion = -1
		}
	}

	if retryAfter != "" {
		// Suggest decrease concurrency if Retry-After is specified
		suggestion = -1
	} else {
		// Suggest increase concurrency if currently below maximum limit and no other decrease suggestion has been made
		if len(ch.sem) < MaxConcurrency && suggestion == 0 {
			suggestion = 1
		}
	}

	return suggestion
}

// MonitorServerResponseCodes monitors the response status codes and suggests a concurrency adjustment.
func (ch *ConcurrencyHandler) MonitorServerResponseCodes(resp *http.Response) int {
	statusCode := resp.StatusCode

	// Lock the metrics to ensure thread safety
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Update the appropriate error count based on the response status code
	switch {
	case statusCode >= 500 && statusCode < 600:
		ch.Metrics.TotalRateLimitErrors++
	case statusCode >= 400 && statusCode < 500:
		// Assuming 4xx errors as client errors
		ch.Metrics.TotalRetries++
	}

	// Calculate error rate
	totalRequests := float64(ch.Metrics.TotalRequests)
	totalErrors := float64(ch.Metrics.TotalRateLimitErrors + ch.Metrics.TotalRetries)
	errorRate := totalErrors / totalRequests

	// Set the new error rate in the metrics
	ch.Metrics.ResponseCodeMetrics.ErrorRate = errorRate

	// Determine action based on the error rate
	if errorRate > ErrorRateThreshold {
		// Suggest decrease concurrency
		return -1
	} else if errorRate <= ErrorRateThreshold && len(ch.sem) < MaxConcurrency {
		// Suggest increase concurrency if there is capacity
		return 1
	}
	return 0
}

// MonitorResponseTimeVariability monitors the response time variability and suggests a concurrency adjustment.
func (ch *ConcurrencyHandler) MonitorResponseTimeVariability(responseTime time.Duration) int {
	ch.Metrics.Lock.Lock()
	defer ch.Metrics.Lock.Unlock()

	// Update ResponseTimeVariability metrics
	ch.Metrics.ResponseTimeVariability.Lock.Lock()
	defer ch.Metrics.ResponseTimeVariability.Lock.Unlock()
	ch.Metrics.ResponseTimeVariability.Total += responseTime
	ch.Metrics.ResponseTimeVariability.Count++

	// Calculate average response time
	ch.Metrics.ResponseTimeVariability.Average = ch.Metrics.ResponseTimeVariability.Total / time.Duration(ch.Metrics.ResponseTimeVariability.Count)

	// Calculate variance of response times
	ch.Metrics.ResponseTimeVariability.Variance = ch.calculateVariance(ch.Metrics.ResponseTimeVariability.Average, responseTime)

	// Calculate standard deviation of response times
	stdDev := math.Sqrt(ch.Metrics.ResponseTimeVariability.Variance)

	// Determine action based on standard deviation
	if stdDev > ch.Metrics.ResponseTimeVariability.StdDevThreshold {
		// Suggest decrease concurrency
		return -1
	} else if stdDev <= ch.Metrics.ResponseTimeVariability.StdDevThreshold && len(ch.sem) < MaxConcurrency {
		// Suggest increase concurrency if there is capacity
		return 1
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
