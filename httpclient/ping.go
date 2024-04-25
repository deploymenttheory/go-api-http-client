// httpclient_ping.go
package httpclient

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/deploymenttheory/go-api-http-client/ratehandler"

	"go.uber.org/zap"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// DoPole performs an HTTP "ping" to the specified endpoint using the given HTTP method, body,
// and output variable. It attempts the request until a 200 OK response is received or the
// maximum number of retry attempts is reached. The function uses a backoff strategy for retries
// to manage load on the server and network. This function is useful for checking the availability
// or health of an endpoint, particularly in environments where network reliability might be an issue.

// Parameters:
// - method: The HTTP method to be used for the request. This should typically be "GET" for a ping operation, but the function is designed to accommodate any HTTP method.
// - endpoint: The target API endpoint for the ping request. This should be a relative path that will be appended to the base URL configured for the HTTP client.
// - body: The payload for the request, if any. For a typical ping operation, this would be nil, but the function is designed to accommodate requests that might require a body.
// - out: A pointer to an output variable where the response will be deserialized. This is included to maintain consistency with the signature of other request functions, but for a ping operation, it might not be used.

// Returns:
// - *http.Response: The HTTP response from the server. In case of a successful ping (200 OK),
// this response contains the status code, headers, and body of the response. In case of errors
// or if the maximum number of retries is reached without a successful response, this will be the
// last received HTTP response.
//
// - error: An error object indicating failure during the execution of the ping operation. This
// could be due to network issues, server errors, or reaching the maximum number of retry attempts
// without receiving a 200 OK response.

// Usage:
// This function is intended for use in scenarios where it's necessary to confirm the availability
// or health of an endpoint, with built-in retry logic to handle transient network or server issues.
// The caller is responsible for handling the response and error according to their needs, including
// closing the response body when applicable to avoid resource leaks.

// Example:
// var result MyResponseType
// resp, err := client.DoPing("GET", "/api/health", nil, &result)
//
//	if err != nil {
//	    // Handle error
//	}
//
// // Process response
func (c *Client) DoPole(method, endpoint string, body, out interface{}) (*http.Response, error) {
	log := c.Logger
	log.Debug("Starting HTTP Ping", zap.String("method", method), zap.String("endpoint", endpoint))

	// Initialize retry count and define maximum retries
	var retryCount int
	maxRetries := c.clientConfig.ClientOptions.Retry.MaxRetryAttempts

	// Loop until a successful response is received or maximum retries are reached
	for retryCount <= maxRetries {
		// Use the existing 'do' function for sending the request
		resp, err := c.executeRequestWithRetries(method, endpoint, body, out)

		// If request is successful and returns 200 status code, return the response
		if err == nil && resp.StatusCode == http.StatusOK {
			log.Debug("Ping successful", zap.String("method", method), zap.String("endpoint", endpoint))
			return resp, nil
		}

		// Increment retry count and log the attempt
		retryCount++
		log.Warn("Ping failed, retrying...", zap.String("method", method), zap.String("endpoint", endpoint), zap.Int("retryCount", retryCount))

		// Calculate backoff duration and wait before retrying
		backoffDuration := ratehandler.CalculateBackoff(retryCount)
		time.Sleep(backoffDuration)
	}

	// If maximum retries are reached without a successful response, return an error
	log.Error("Ping failed after maximum retries", zap.String("method", method), zap.String("endpoint", endpoint))
	return nil, fmt.Errorf("ping failed after %d retries", maxRetries)
}

// DoPing performs an ICMP "ping" to the specified host. It sends ICMP echo requests and waits for echo replies.
// This function is useful for checking the availability or health of a host, particularly in environments where
// network reliability might be an issue.

// Parameters:
// - host: The target host for the ping request.
// - timeout: The timeout for waiting for a ping response.

// Returns:
// - error: An error object indicating failure during the execution of the ping operation or nil if the ping was successful.

// Usage:
// This function is intended for use in scenarios where it's necessary to confirm the availability or health of a host.
// The caller is responsible for handling the error according to their needs.

// Example:
// err := client.DoPing("www.example.com", 3*time.Second)
// if err != nil {
//     // Handle error
// }

func (c *Client) DoPing(host string, timeout time.Duration) error {
	log := c.Logger

	// Listen for ICMP replies
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		log.Error("Failed to listen for ICMP packets", zap.Error(err))
		return fmt.Errorf("failed to listen for ICMP packets: %w", err)
	}
	defer conn.Close()

	// Resolve the IP address of the host
	dst, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		log.Error("Failed to resolve IP address", zap.String("host", host), zap.Error(err))
		return fmt.Errorf("failed to resolve IP address for host %s: %w", host, err)
	}

	// Create an ICMP Echo Request message
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1, // Use PID as ICMP ID
			Data: []byte("HELLO"), // Data payload
		},
	}

	// Marshal the message into bytes
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		log.Error("Failed to marshal ICMP message", zap.Error(err))
		return fmt.Errorf("failed to marshal ICMP message: %w", err)
	}

	// Send the ICMP Echo Request message
	if _, err := conn.WriteTo(msgBytes, dst); err != nil {
		log.Error("Failed to send ICMP message", zap.Error(err))
		return fmt.Errorf("failed to send ICMP message: %w", err)
	}

	// Set read timeout
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		log.Error("Failed to set read deadline", zap.Error(err))
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Wait for an ICMP Echo Reply message
	reply := make([]byte, 1500)
	n, _, err := conn.ReadFrom(reply)
	if err != nil {
		log.Error("Failed to receive ICMP reply", zap.Error(err))
		return fmt.Errorf("failed to receive ICMP reply: %w", err)
	}

	// Parse the ICMP message from the reply
	parsedMsg, err := icmp.ParseMessage(1, reply[:n])
	if err != nil {
		log.Error("Failed to parse ICMP message", zap.Error(err))
		return fmt.Errorf("failed to parse ICMP message: %w", err)
	}

	// Check if the message is an ICMP Echo Reply
	if echoReply, ok := parsedMsg.Type.(*ipv4.ICMPType); ok {
		if *echoReply != ipv4.ICMPTypeEchoReply {
			log.Error("Did not receive ICMP Echo Reply", zap.String("received_type", echoReply.String()))
			return fmt.Errorf("did not receive ICMP Echo Reply, received type: %s", echoReply.String())
		}
	} else {
		// Handle the case where the type assertion fails
		log.Error("Failed to assert ICMP message type")
		return fmt.Errorf("failed to assert ICMP message type")
	}

	log.Info("Ping successful", zap.String("host", host))
	return nil
}
