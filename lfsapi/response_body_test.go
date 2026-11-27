package lfsapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimitedWithPlainTextBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limit exceeded by gateway"))
	}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	c := NewClient(nil)
	_, err = c.Do(req)
	require.EqualError(t, err, "LFS: rate limit exceeded by gateway")
}
