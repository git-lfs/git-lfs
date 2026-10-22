package lfshttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"testing/iotest"

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
	c := NewClient(nil)
	c.Verbose = true
	c.VerboseOut = out

	req, err := http.NewRequest("POST", srv.URL, nil)
	req.Header.Set("Authorization", "Basic ABC")
	req.Header.Set("Content-Type", "application/json")
	require.Nil(t, err)
	require.Nil(t, MarshalToRequest(req, verboseTest{"Verbose"}))

	res, err := c.Do(req)
	require.Nil(t, err)
	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)

	s := out.String()
	t.Log(s)

	expected := []string{
		"> Host: 127.0.0.1:",
		"\n> Authorization: Basic * * * * *\n",
		"\n> Content-Type: application/json\n",
		"\n> Accept-Encoding: gzip\n",
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
		by, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		assert.Equal(t, "binary-request", string(by))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte(`binary-response`))
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	c := NewClient(nil)
	c.Verbose = true
	c.VerboseOut = out

	buf := bytes.NewBufferString("binary-request")
	req, err := http.NewRequest("POST", srv.URL, buf)
	req.Header.Set("Authorization", "Basic ABC")
	req.Header.Set("Content-Type", "application/octet-stream")
	require.Nil(t, err)

	res, err := c.Do(req)
	require.Nil(t, err)
	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)

	s := out.String()
	t.Log(s)

	expected := []string{
		"> Host: 127.0.0.1:",
		"\n> Authorization: Basic * * * * *\n",
		"\n> Content-Type: application/octet-stream\n",
		"\n> Accept-Encoding: gzip\n",

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
	c := NewClient(nil)
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
	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)

	s := out.String()
	t.Log(s)

	expected := []string{
		"> Host: 127.0.0.1:",
		"\n> Authorization: Basic ABC\n",
		"\n> Content-Type: application/json\n",
		"\n> Accept-Encoding: gzip\n",
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
	c := NewClient(nil)
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
	io.Copy(io.Discard, res.Body)
	res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 1, called)
	assert.EqualValues(t, 0, out.Len(), out.String())
}

func TestVerboseEnabledWithAuthorizationInBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
		w.Write([]byte(`{"objects":[{"actions":{"download":{"href":"https://example.com","header":{"Authorization":"Basic dXNlcjpwYXNz"}}}}]}`))
	}))
	defer srv.Close()

	readers := map[string]func(io.Reader) io.Reader{
		"full reads":     func(r io.Reader) io.Reader { return r },
		"one-byte reads": iotest.OneByteReader,
		"half reads":     iotest.HalfReader,
		"data with EOF":  iotest.DataErrReader,
	}

	for desc, reader := range readers {
		t.Run(desc, func(t *testing.T) {
			out := &bytes.Buffer{}
			c := NewClient(nil)
			c.Verbose = true
			c.VerboseOut = out

			req, err := http.NewRequest("GET", srv.URL, nil)
			require.Nil(t, err)

			res, err := c.Do(req)
			require.Nil(t, err)
			io.Copy(io.Discard, reader(res.Body))
			res.Body.Close()

			assert.Equal(t, 200, res.StatusCode)

			s := out.String()
			t.Log(s)

			assert.True(t, strings.Contains(s, "\n< \n"+`{"objects":[{"actions":{"download":{"href":"https://example.com","header":{"Authorization":"Basic * * * * *"}}}}]}`))
			assert.False(t, strings.Contains(s, "dXNlcjpwYXNz"), "contains authorization credentials")
		})
	}
}

func TestVerboseEnabledWithDebuggingAndAuthorizationInBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.git-lfs+json")
		w.Write([]byte(`{"objects":[{"actions":{"download":{"href":"https://example.com","header":{"Authorization":"Basic dXNlcjpwYXNz"}}}}]}`))
	}))
	defer srv.Close()

	out := &bytes.Buffer{}
	c := NewClient(nil)
	c.Verbose = true
	c.VerboseOut = out
	c.DebuggingVerbose = true

	req, err := http.NewRequest("GET", srv.URL, nil)
	require.Nil(t, err)

	res, err := c.Do(req)
	require.Nil(t, err)
	io.Copy(io.Discard, iotest.OneByteReader(res.Body))
	res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)

	s := out.String()
	t.Log(s)

	assert.True(t, strings.Contains(s, "\n< \n"+`{"objects":[{"actions":{"download":{"href":"https://example.com","header":{"Authorization":"Basic dXNlcjpwYXNz"}}}}]}`))
}

func TestAuthRedactor(t *testing.T) {
	for desc, c := range map[string]struct {
		Input    string
		Expected string
	}{
		"basic": {
			`{"header":{"Authorization":"Basic dXNlcjpwYXNz"}}`,
			`{"header":{"Authorization":"Basic * * * * *"}}`,
		},
		"lowercase with whitespace": {
			`{"header": {"authorization" : "Bearer dG9rZW4="}}`,
			`{"header": {"authorization" : "Bearer * * * * *"}}`,
		},
		"no scheme": {
			`{"header":{"Authorization":"dG9rZW4="}}`,
			`{"header":{"Authorization":"* * * * *"}}`,
		},
		"escaped quotes": {
			`{"header":{"Authorization":"Signature sig=\"c2ln\""}}`,
			`{"header":{"Authorization":"Signature * * * * *"}}`,
		},
		"multiple": {
			`[{"Authorization":"Basic dXNlcjpwYXNz"},{"Authorization":"RemoteAuth dG9rZW4="}]`,
			`[{"Authorization":"Basic * * * * *"},{"Authorization":"RemoteAuth * * * * *"}]`,
		},
		"other uses": {
			`{"type":"authorization","message":"authorization required","authorization":{"x":"y"}}`,
			`{"type":"authorization","message":"authorization required","authorization":{"x":"y"}}`,
		},
		"truncated value": {
			`{"header":{"Authorization":"Basic dXNlcjpw`,
			`{"header":{"Authorization":"Basic * * * * *`,
		},
		"truncated key": {
			`{"header":{"Author`,
			`{"header":{"Author`,
		},
	} {
		t.Run(desc, func(t *testing.T) {
			// Split the input into two reads at every possible
			// position.
			for i := 0; i <= len(c.Input); i++ {
				a := &authRedactor{}
				out := string(a.redact([]byte(c.Input[:i]), false))
				assert.True(t, strings.HasPrefix(c.Expected, out), "split at %d: %q", i, out)
				out += string(a.redact([]byte(c.Input[i:]), false))
				out += string(a.redact(nil, true))
				assert.Equal(t, c.Expected, out, "split at %d", i)
			}

			// Read the input one byte at a time, so the output
			// traced so far must never contain any credentials.
			a := &authRedactor{}
			var out string
			for i := 0; i < len(c.Input); i++ {
				out += string(a.redact([]byte(c.Input[i:i+1]), false))
				assert.True(t, strings.HasPrefix(c.Expected, out), "read %d: %q", i, out)
			}
			out += string(a.redact(nil, true))
			assert.Equal(t, c.Expected, out)
		})
	}
}

func TestAuthRedactorWithLongValue(t *testing.T) {
	a := &authRedactor{}
	out := string(a.redact([]byte(`{"header":{"Authorization":"Basic `), false))
	assert.Equal(t, `{"header":{`, out)

	value := bytes.Repeat([]byte("A"), authRedactorMaxWithheld)
	out = string(a.redact(value, false))
	assert.Equal(t, `"Authorization":"Basic * * * * *`, out)
	assert.Empty(t, a.withheld)
}
