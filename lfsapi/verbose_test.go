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

type verboseTest struct {
	Test string
}

func TestVerboseEnabled(t *testing.T) {
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
	c.Verbose = true
	c.VerboseOut = out

	req, err := http.NewRequest("POST", srv.URL, nil)
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

	s := out.String()
	t.Log(s)

	expected := []string{
		"> Host: 127.0.0.1:",
		"\n> Authorization: Basic * * * * *\n",
		"\n> Content-Type: application/json\n",
		"\n> \n" + `{"Test":"Verbose"}` + "\n\n",

		"\n< HTTP/1.1 200 OK\n",
		"\n< Content-Type: application/json\n",
		"\n< \n" + `{"Status":"Ok"}`,
	}

	for _, substr := range expected {
		if !assert.True(t, strings.Contains(s, substr)) {
			t.Logf("missing: %q", substr)
		}
	}
}

func TestVerboseWithBinaryBody(t *testing.T) {
	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called, 1)
		t.Logf("srv req %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		assert.Equal(t, "Basic ABC", r.Header.Get("Authorization"))
		by, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		assert.Equal(t, "binary-request", string(by))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(`binary-response`))
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	c, _ := NewClient(nil)
	c.Verbose = true
	c.VerboseOut = out

	buf := bytes.NewBufferString("binary-request")
	req, err := http.NewRequest("POST", srv.URL, buf)
	req.Header.Set("Authorization", "Basic ABC")
	req.Header.Set("Content-Type", "application/octet-stream")
	require.Nil(t, err)

	res, err := c.Do(req)
	require.Nil(t, err)
	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)

	s := out.String()
	t.Log(s)

	expected := []string{
		"> Host: 127.0.0.1:",
		"\n> Authorization: Basic * * * * *\n",
		"\n> Content-Type: application/octet-stream\n",

		"\n< HTTP/1.1 200 OK\n",
		"\n< Content-Type: application/octet-stream\n",
	}

	for _, substr := range expected {
		if !assert.True(t, strings.Contains(s, substr)) {
			t.Logf("missing: %q", substr)
		}
	}

	assert.False(t, strings.Contains(s, "binary"), "contains binary request or response body")
}

func TestVerboseEnabledWithDebugging(t *testing.T) {
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
	c.Verbose = true
	c.VerboseOut = out
	c.DebuggingVerbose = true

	req, err := http.NewRequest("POST", srv.URL, nil)
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

	s := out.String()
	t.Log(s)

	expected := []string{
		"> Host: 127.0.0.1:",
		"\n> Authorization: Basic ABC\n",
		"\n> Content-Type: application/json\n",
		"\n> \n" + `{"Test":"Verbose"}` + "\n\n",

		"\n< HTTP/1.1 200 OK\n",
		"\n< Content-Type: application/json\n",
		"\n< \n" + `{"Status":"Ok"}`,
	}

	for _, substr := range expected {
		if !assert.True(t, strings.Contains(s, substr)) {
			t.Logf("missing: %q", substr)
		}
	}
}

func TestVerboseDisabled(t *testing.T) {
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
	c.Verbose = false
	c.VerboseOut = out
	c.DebuggingVerbose = true

	req, err := http.NewRequest("POST", srv.URL, nil)
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
	assert.EqualValues(t, 0, out.Len(), out.String())
}
