package httpclient

import (
	"time"
)

// Amends the HTTP timeout time
func (c *Client) ModifyHttpTimeout(newTimeout time.Duration) {
	c.http.Timeout = newTimeout
}

// Resets HTTP timeout time back to 10 seconds
func (c *Client) ResetTimeout() {
	c.http.Timeout = DefaultTimeout
}

func (c *Client) HttpTimeout() time.Duration {
	return c.http.Timeout
}
