package lfshttp

import (
	"context"
	"strings"
	"time"

	"github.com/rubyist/tracerx"
)

const defaultMaxRetryTime = 300

func (c *Client) setRetryAfter(hostname string, retryAt time.Time) {
	hostname = strings.ToLower(hostname)
	delay := time.Until(retryAt)
	// TransferQueue rejects server-requested delays beyond maxRetryTime. Do
	// not stall otherwise viable requests to this host for a delay the queue
	// will not honor.
	if hostname == "" || delay <= 0 || delay > c.maxRetryTime() {
		return
	}

	c.retryMu.Lock()
	defer c.retryMu.Unlock()

	if current := c.retryAfter[hostname]; !retryAt.After(current) {
		return
	}
	if c.retryAfter == nil {
		c.retryAfter = make(map[string]time.Time)
	}
	c.retryAfter[hostname] = retryAt
}

func (c *Client) maxRetryTime() time.Duration {
	seconds := defaultMaxRetryTime
	if c.gitEnv != nil {
		if configured := c.gitEnv.Int("lfs.transfer.maxretrytime", 0); configured > 0 {
			seconds = configured
		}
	}
	return time.Duration(seconds) * time.Second
}

func (c *Client) waitForRetryAfter(ctx context.Context, hostname string) error {
	hostname = strings.ToLower(hostname)
	if hostname == "" {
		return nil
	}

	for {
		c.retryMu.Lock()
		retryAt, ok := c.retryAfter[hostname]
		wait := time.Until(retryAt)
		if !ok || wait <= 0 {
			if ok {
				delete(c.retryAfter, hostname)
			}
			c.retryMu.Unlock()
			return nil
		}
		c.retryMu.Unlock()

		tracerx.Printf("http: waiting %.2fs before sending another request to %s", wait.Seconds(), hostname)
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return ctx.Err()
		}
	}
}
