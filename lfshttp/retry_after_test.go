package lfshttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientWaitsForRetryAfterOnSameHost(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/rate-limited" {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(nil)
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/rate-limited", nil)
	require.NoError(t, err)
	res, err := c.Do(req)
	require.Error(t, err)
	require.Equal(t, http.StatusTooManyRequests, res.StatusCode)
	require.NoError(t, res.Body.Close())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/next", nil)
	require.NoError(t, err)
	_, err = c.Do(req)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.EqualValues(t, 1, requests.Load(), "request should be cancelled before reaching the rate-limited host")

	started := time.Now()
	req, err = http.NewRequest(http.MethodGet, srv.URL+"/next", nil)
	require.NoError(t, err)
	res, err = c.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	assert.GreaterOrEqual(t, time.Since(started), 850*time.Millisecond)
	assert.EqualValues(t, 2, requests.Load())
}

func TestClientRetryAfterIsScopedByHostname(t *testing.T) {
	c := NewClient(nil)
	c.setRetryAfter("limited.example.com", time.Now().Add(time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	require.NoError(t, c.waitForRetryAfter(ctx, "other.example.com"))

	err := c.waitForRetryAfter(ctx, "LIMITED.EXAMPLE.COM")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestClientRetryAfterKeepsLatestDeadline(t *testing.T) {
	c := NewClient(nil)
	hostname := "limited.example.com"
	later := time.Now().Add(2 * time.Second)
	c.setRetryAfter(hostname, later)
	c.setRetryAfter(hostname, later.Add(-time.Second))

	c.retryMu.Lock()
	actual := c.retryAfter[hostname]
	c.retryMu.Unlock()
	assert.Equal(t, later, actual)
}

func TestClientDoesNotWaitPastMaxRetryTime(t *testing.T) {
	c := NewClient(NewContext(nil, nil, map[string]string{
		"lfs.transfer.maxretrytime": "1",
	}))
	c.setRetryAfter("limited.example.com", time.Now().Add(2*time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	require.NoError(t, c.waitForRetryAfter(ctx, "limited.example.com"))
}
