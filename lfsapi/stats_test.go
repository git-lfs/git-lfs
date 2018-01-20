package lfsapi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsWithKey(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called, 1)
		t.Logf("srv req %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		assert.Equal(t, "Basic ABC", r.Header.Get("Authorization"))
		body := &verboseTest{}
		err := json.NewDecoder(r.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Verbose", body.Test)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Status":"Ok"}`))
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	c, _ := NewClient(nil)
	c.ConcurrentTransfers = 5
	c.LogHTTPStats(nopCloser(out))

	req, err := http.NewRequest("POST", srv.URL, nil)
	req = c.LogRequest(req, "stats-test")
	req.Header.Set("Authorization", "Basic ABC")
	req.Header.Set("Content-Type", "application/json")
	require.Nil(t, err)
	require.Nil(t, MarshalToRequest(req, verboseTest{"Verbose"}))

	res, err := c.Do(req)
	require.Nil(t, err)

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()
	assert.Nil(t, c.Close())

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)

	stats := strings.TrimSpace(out.String())
	t.Log(stats)
	lines := strings.Split(stats, "\n")
	require.Equal(t, 3, len(lines))
	assert.True(t, strings.Contains(lines[0], "concurrent=5"))
	expected := []string{
		"key=stats-test",
		"event=request",
		"body=18",
		"url=" + srv.URL,
	}

	for _, substr := range expected {
		assert.True(t, strings.Contains(lines[1], substr), "missing: "+substr)
	}

	expected = []string{
		"key=stats-test",
		"event=response",
		"url=" + srv.URL,
	}

	for _, substr := range expected {
		assert.True(t, strings.Contains(lines[2], substr), "missing: "+substr)
	}
}

func TestStatsWithoutKey(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called, 1)
		t.Logf("srv req %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		assert.Equal(t, "Basic ABC", r.Header.Get("Authorization"))
		body := &verboseTest{}
		err := json.NewDecoder(r.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Verbose", body.Test)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Status":"Ok"}`))
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	c, _ := NewClient(nil)
	c.ConcurrentTransfers = 5
	c.LogHTTPStats(nopCloser(out))

	req, err := http.NewRequest("POST", srv.URL, nil)
	req.Header.Set("Authorization", "Basic ABC")
	req.Header.Set("Content-Type", "application/json")
	require.Nil(t, err)
	require.Nil(t, MarshalToRequest(req, verboseTest{"Verbose"}))

	res, err := c.Do(req)
	require.Nil(t, err)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()
	assert.Nil(t, c.Close())

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)

	stats := strings.TrimSpace(out.String())
	t.Log(stats)
	assert.True(t, strings.Contains(stats, "concurrent=5"))
	assert.Equal(t, 1, len(strings.Split(stats, "\n")))
}

func TestStatsDisabled(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called, 1)
		t.Logf("srv req %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		assert.Equal(t, "Basic ABC", r.Header.Get("Authorization"))
		body := &verboseTest{}
		err := json.NewDecoder(r.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Verbose", body.Test)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Status":"Ok"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(nil)
	c.ConcurrentTransfers = 5

	req, err := http.NewRequest("POST", srv.URL, nil)
	req = c.LogRequest(req, "stats-test")

	req.Header.Set("Authorization", "Basic ABC")
	req.Header.Set("Content-Type", "application/json")
	require.Nil(t, err)
	require.Nil(t, MarshalToRequest(req, verboseTest{"Verbose"}))

	res, err := c.Do(req)
	require.Nil(t, err)

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)

	out := &bytes.Buffer{}
	c.LogStats(out)
	assert.Equal(t, 0, out.Len())
}

func nopCloser(w io.Writer) io.WriteCloser {
	return nopWCloser{w}
}

type nopWCloser struct {
	io.Writer
}

func (w nopWCloser) Close() error {
	return nil
}
