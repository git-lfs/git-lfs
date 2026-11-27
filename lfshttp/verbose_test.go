package lfshttp

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/v3/errors"
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

func TestTraceTLS(t *testing.T) {
	req, err := http.NewRequest("GET", "https://lfs.example.com/objects/batch", nil)
	require.NoError(t, err)

	c := NewClient(nil)

	// request comes back untouched,
	// since c.TraceTLS is false
	assert.Same(t, req, c.traceTLS(req))
	assert.Nil(t, httptrace.ContextClientTrace(req.Context()))

	c.TraceTLS = true
	traced := c.traceTLS(req)

	// client trace should exist,
	// since c.TraceTLS is set to true
	trace := httptrace.ContextClientTrace(traced.Context())
	require.NotNil(t, trace)
	assert.NotNil(t, trace.TLSHandshakeDone)

	// already traced request stays the same
	assert.Same(t, traced, c.traceTLS(traced))
}

func TestTraceTLSState(t *testing.T) {
	leaf := &x509.Certificate{
		Subject:            pkix.Name{CommonName: "lfs.example.com", Organization: []string{"Git LFS"}},
		Issuer:             pkix.Name{CommonName: "LFS Test CA", Organization: []string{"Git LFS"}},
		PublicKeyAlgorithm: x509.RSA,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	root := &x509.Certificate{
		Subject: pkix.Name{CommonName: "LFS Test CA", Organization: []string{"Git LFS"}},
	}

	verified := tls.ConnectionState{
		Version:            tls.VersionTLS13,
		CipherSuite:        tls.TLS_AES_128_GCM_SHA256,
		NegotiatedProtocol: "h2",
		PeerCertificates:   []*x509.Certificate{leaf, root},
		VerifiedChains:     [][]*x509.Certificate{{leaf, root}},
	}

	unverified := verified
	unverified.VerifiedChains = nil

	resumed := verified
	resumed.DidResume = true

	const negotiated = `negotiated TLS 1.3 / TLS_AES_128_GCM_SHA256 (ALPN "h2")`

	certLines := []string{
		"server certificate subject: CN=lfs.example.com,O=Git LFS",
		"server certificate issuer: CN=LFS Test CA,O=Git LFS",
		"server certificate algorithms: RSA / SHA256-RSA",
		"server certificate chain: CN=lfs.example.com,O=Git LFS -> CN=LFS Test CA,O=Git LFS",
	}

	tests := []struct {
		name     string
		state    tls.ConnectionState
		err      error
		expected []string
	}{
		{
			name:  "verified certificate",
			state: verified,
			expected: slices.Concat(
				[]string{negotiated},
				certLines,
				[]string{"server certificate verified"},
			),
		},
		{
			name:  "unverified certificate",
			state: unverified,
			expected: slices.Concat(
				[]string{negotiated},
				certLines,
				[]string{"server certificate not verified"},
			),
		},
		{
			name:  "resumed connection",
			state: resumed,
			expected: slices.Concat(
				[]string{negotiated, "connection resumed"},
				certLines,
				[]string{"server certificate verified"},
			),
		},
		{
			name: "no peer certificates",
			state: tls.ConnectionState{
				Version:     tls.VersionTLS12,
				CipherSuite: tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
			expected: []string{"negotiated TLS 1.2 / TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"},
		},
		{
			name:     "unknown cipher suite",
			state:    tls.ConnectionState{Version: tls.VersionTLS13, CipherSuite: 0xabcd, NegotiatedProtocol: "h2"},
			expected: []string{`negotiated TLS 1.3 / 0xABCD (ALPN "h2")`},
		},
		{
			name:     "handshake failed",
			err:      errors.New("x509: certificate signed by unknown authority"),
			expected: []string{"handshake failed: x509: certificate signed by unknown authority"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, traceTLSState(tt.state, tt.err))
		})
	}
}

func TestTraceTLSRequestSucceeds(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"Status": "ok"}`))
	}))
	defer srv.Close()

	c := NewClient(nil)
	c.TraceTLS = true
	c.SkipSSLVerify = true // httptest uses self-signed

	req, err := http.NewRequest("GET", srv.URL, nil)
	require.Nil(t, err)

	response, err := c.Do(req)
	require.Nil(t, err)

	io.Copy(io.Discard, response.Body)
	response.Body.Close()

	assert.Equal(t, 200, response.StatusCode)
}
