package httpclient

import (
	"sync"
	"time"
)

var mu sync.Mutex

// Amends the HTTP timeout time
func (c *Client) ModifyHttpTimeout(newTimeout time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	c.http.Timeout = newTimeout
}

// Resets HTTP timeout time back to 10 seconds
func (c *Client) ResetTimeout() {
	mu.Lock()
	defer mu.Unlock()
	c.http.Timeout = DefaultTimeout
}
