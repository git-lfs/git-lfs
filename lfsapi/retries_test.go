package lfsapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRetries(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req = WithRetries(req, 1)

	n, ok := Retries(req)
	assert.True(t, ok)
	assert.Equal(t, 1, n)
}

func TestRetriesOnUnannotatedRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	n, ok := Retries(req)
	assert.False(t, ok)
	assert.Equal(t, 0, n)
}

func TestRequestWithRetries(t *testing.T) {
	type T struct {
		S string `json:"s"`
	}

	var hasRaw bool = true
	var requests uint32
	var berr error

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload T
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			berr = err
		}

		assert.Equal(t, "Hello, world!", payload.S)

		if atomic.AddUint32(&requests, 1) < 3 {
			raw, ok := w.(http.Hijacker)
			if !ok {
				hasRaw = false
				return
			}

			conn, _, err := raw.Hijack()
			require.NoError(t, err)
			require.NoError(t, conn.Close())
			return
		}
	}))
	defer srv.Close()

	c, err := NewClient(nil)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", srv.URL, nil)
	require.NoError(t, err)
	require.NoError(t, MarshalToRequest(req, &T{"Hello, world!"}))

	if !hasRaw {
		// Skip tests where the implementation of
		// net/http/httptest.Server does not provide raw access to the
		// connection.
		//
		// Defer the skip outside of the server, since t.Skip halts the
		// running goroutine.
		t.Skip("lfsapi: net/http/httptest.Server does not provide raw access")
	}

	res, err := c.Do(WithRetries(req, 8))
	assert.NoError(t, berr)
	assert.NoError(t, err)
	require.NotNil(t, res, "lfsapi: expected response")

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
