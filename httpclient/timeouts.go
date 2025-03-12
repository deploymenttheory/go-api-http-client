package httpclient

import (
	"sync"
	"time"
)

var mu sync.Mutex

// Modifies the HTTP timeout time
func (c *Client) ModifyHttpTimeout(newTimeout time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	c.http.Timeout = newTimeout
}
