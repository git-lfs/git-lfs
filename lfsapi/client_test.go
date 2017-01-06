package lfsapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type redirectTest struct {
	Test string
}

func TestClientRedirect(t *testing.T) {
	var called1 uint32
	var called2 uint32
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called2, 1)
		t.Logf("srv2 req %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		switch r.URL.Path {
		case "/ok":
			assert.Equal(t, "", r.Header.Get("Authorization"))
			assert.Equal(t, "1", r.Header.Get("A"))
			body := &redirectTest{}
			err := json.NewDecoder(r.Body).Decode(body)
			assert.Nil(t, err)
			assert.Equal(t, "External", body.Test)

			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
		}
	}))

	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called1, 1)
		t.Logf("srv1 req %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		switch r.URL.Path {
		case "/local":
			w.Header().Set("Location", "/ok")
			w.WriteHeader(307)
		case "/external":
			w.Header().Set("Location", srv2.URL+"/ok")
			w.WriteHeader(307)
		case "/ok":
			assert.Equal(t, "auth", r.Header.Get("Authorization"))
			assert.Equal(t, "1", r.Header.Get("A"))
			body := &redirectTest{}
			err := json.NewDecoder(r.Body).Decode(body)
			assert.Nil(t, err)
			assert.Equal(t, "Local", body.Test)

			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv1.Close()
	defer srv2.Close()

	c := &Client{}

	// local redirect
	req, err := http.NewRequest("POST", srv1.URL+"/local", nil)
	require.Nil(t, err)
	req.Header.Set("Authorization", "auth")
	req.Header.Set("A", "1")

	require.Nil(t, MarshalToRequest(req, &redirectTest{Test: "Local"}))

	res, err := c.Do(req)
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 2, called1)
	assert.EqualValues(t, 0, called2)

	// external redirect
	req, err = http.NewRequest("POST", srv1.URL+"/external", nil)
	require.Nil(t, err)
	req.Header.Set("Authorization", "auth")
	req.Header.Set("A", "1")

	require.Nil(t, MarshalToRequest(req, &redirectTest{Test: "External"}))

	res, err = c.Do(req)
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 3, called1)
	assert.EqualValues(t, 1, called2)
}

func TestNewClient(t *testing.T) {
	c, err := NewClient(TestEnv(map[string]string{}), TestEnv(map[string]string{
		"lfs.dialtimeout":         "151",
		"lfs.keepalive":           "152",
		"lfs.tlstimeout":          "153",
		"lfs.concurrenttransfers": "154",
	}))

	require.Nil(t, err)
	assert.Equal(t, 151, c.DialTimeout)
	assert.Equal(t, 152, c.KeepaliveTimeout)
	assert.Equal(t, 153, c.TLSTimeout)
	assert.Equal(t, 154, c.ConcurrentTransfers)
}

func TestNewClientWithGitSSLVerify(t *testing.T) {
	c, err := NewClient(nil, nil)
	assert.Nil(t, err)
	assert.False(t, c.SkipSSLVerify)

	for _, value := range []string{"true", "1", "t"} {
		c, err = NewClient(TestEnv(map[string]string{}), TestEnv(map[string]string{
			"http.sslverify": value,
		}))
		t.Logf("http.sslverify: %q", value)
		assert.Nil(t, err)
		assert.False(t, c.SkipSSLVerify)
	}

	for _, value := range []string{"false", "0", "f"} {
		c, err = NewClient(TestEnv(map[string]string{}), TestEnv(map[string]string{
			"http.sslverify": value,
		}))
		t.Logf("http.sslverify: %q", value)
		assert.Nil(t, err)
		assert.True(t, c.SkipSSLVerify)
	}
}

func TestNewClientWithOSSSLVerify(t *testing.T) {
	c, err := NewClient(nil, nil)
	assert.Nil(t, err)
	assert.False(t, c.SkipSSLVerify)

	for _, value := range []string{"false", "0", "f"} {
		c, err = NewClient(TestEnv(map[string]string{
			"GIT_SSL_NO_VERIFY": value,
		}), TestEnv(map[string]string{}))
		t.Logf("GIT_SSL_NO_VERIFY: %q", value)
		assert.Nil(t, err)
		assert.False(t, c.SkipSSLVerify)
	}

	for _, value := range []string{"true", "1", "t"} {
		c, err = NewClient(TestEnv(map[string]string{
			"GIT_SSL_NO_VERIFY": value,
		}), TestEnv(map[string]string{}))
		t.Logf("GIT_SSL_NO_VERIFY: %q", value)
		assert.Nil(t, err)
		assert.True(t, c.SkipSSLVerify)
	}
}

func TestNewRequest(t *testing.T) {
	tests := [][]string{
		{"https://example.com", "a", "https://example.com/a"},
		{"https://example.com/", "a", "https://example.com/a"},
		{"https://example.com/a", "b", "https://example.com/a/b"},
		{"https://example.com/a/", "b", "https://example.com/a/b"},
	}

	for _, test := range tests {
		c, err := NewClient(nil, TestEnv(map[string]string{
			"lfs.url": test[0],
		}))
		require.Nil(t, err)

		req, err := c.NewRequest("POST", c.Endpoints.Endpoint("", ""), test[1], nil)
		require.Nil(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, test[2], req.URL.String(), fmt.Sprintf("endpoint: %s, suffix: %s, expected: %s", test[0], test[1], test[2]))
	}
}
