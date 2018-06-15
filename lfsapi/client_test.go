package lfsapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type redirectTest struct {
	Test string
}

func TestClientRedirect(t *testing.T) {
	var srv3Https, srv3Http string

	var called1 uint32
	var called2 uint32
	var called3 uint32
	srv3 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called3, 1)
		t.Logf("srv3 req %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		switch r.URL.Path {
		case "/upgrade":
			assert.Equal(t, "auth", r.Header.Get("Authorization"))
			assert.Equal(t, "1", r.Header.Get("A"))
			w.Header().Set("Location", srv3Https+"/upgraded")
			w.WriteHeader(301)
		case "/upgraded":
			// Since srv3 listens on both a TLS-enabled socket and a
			// TLS-disabled one, they are two different hosts.
			// Ensure that, even though this is a "secure" upgrade,
			// the authorization header is stripped.
			assert.Equal(t, "", r.Header.Get("Authorization"))
			assert.Equal(t, "1", r.Header.Get("A"))

		case "/downgrade":
			assert.Equal(t, "auth", r.Header.Get("Authorization"))
			assert.Equal(t, "1", r.Header.Get("A"))
			w.Header().Set("Location", srv3Http+"/404")
			w.WriteHeader(301)

		default:
			w.WriteHeader(404)
		}
	}))

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
	defer srv3.Close()

	srv3InsecureListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.Nil(t, err)

	go http.Serve(srv3InsecureListener, srv3.Config.Handler)
	defer srv3InsecureListener.Close()

	srv3Https = srv3.URL
	srv3Http = fmt.Sprintf("http://%s", srv3InsecureListener.Addr().String())

	c, err := NewClient(NewContext(nil, nil, map[string]string{
		fmt.Sprintf("http.%s.sslverify", srv3Https):  "false",
		fmt.Sprintf("http.%s/.sslverify", srv3Https): "false",
		fmt.Sprintf("http.%s.sslverify", srv3Http):   "false",
		fmt.Sprintf("http.%s/.sslverify", srv3Http):  "false",
		fmt.Sprintf("http.sslverify"):                "false",
	}))
	require.Nil(t, err)

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

	// http -> https (secure upgrade)

	req, err = http.NewRequest("POST", srv3Http+"/upgrade", nil)
	require.Nil(t, err)
	req.Header.Set("Authorization", "auth")
	req.Header.Set("A", "1")

	require.Nil(t, MarshalToRequest(req, &redirectTest{Test: "http->https"}))

	res, err = c.Do(req)
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.EqualValues(t, 2, atomic.LoadUint32(&called3))

	// https -> http (insecure downgrade)

	req, err = http.NewRequest("POST", srv3Https+"/downgrade", nil)
	require.Nil(t, err)
	req.Header.Set("Authorization", "auth")
	req.Header.Set("A", "1")

	require.Nil(t, MarshalToRequest(req, &redirectTest{Test: "https->http"}))

	_, err = c.Do(req)
	assert.EqualError(t, err, "lfsapi/client: refusing insecure redirect, https->http")
}

func TestClientRedirectReauthenticate(t *testing.T) {
	var srv1, srv2 *httptest.Server
	var called1, called2 uint32
	var creds1, creds2 Creds

	srv1 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called1, 1)

		if hdr := r.Header.Get("Authorization"); len(hdr) > 0 {
			parts := strings.SplitN(hdr, " ", 2)
			typ, b64 := parts[0], parts[1]

			auth, err := base64.URLEncoding.DecodeString(b64)
			assert.Nil(t, err)
			assert.Equal(t, "Basic", typ)
			assert.Equal(t, "user1:pass1", string(auth))

			http.Redirect(w, r, srv2.URL+r.URL.Path, http.StatusMovedPermanently)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))

	srv2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&called2, 1)

		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		typ, b64 := parts[0], parts[1]

		auth, err := base64.URLEncoding.DecodeString(b64)
		assert.Nil(t, err)
		assert.Equal(t, "Basic", typ)
		assert.Equal(t, "user2:pass2", string(auth))
	}))

	// Change the URL of srv2 to make it appears as if it is a different
	// host.
	srv2.URL = strings.Replace(srv2.URL, "127.0.0.1", "0.0.0.0", 1)

	creds1 = Creds(map[string]string{
		"protocol": "http",
		"host":     strings.TrimPrefix(srv1.URL, "http://"),

		"username": "user1",
		"password": "pass1",
	})
	creds2 = Creds(map[string]string{
		"protocol": "http",
		"host":     strings.TrimPrefix(srv2.URL, "http://"),

		"username": "user2",
		"password": "pass2",
	})

	defer srv1.Close()
	defer srv2.Close()

	c, err := NewClient(NewContext(nil, nil, nil))
	creds := newCredentialCacher()
	creds.Approve(creds1)
	creds.Approve(creds2)
	c.Credentials = creds

	req, err := http.NewRequest("GET", srv1.URL, nil)
	require.Nil(t, err)

	_, err = c.DoWithAuth("", req)
	assert.Nil(t, err)

	// called1 is 2 since LFS tries an unauthenticated request first
	assert.EqualValues(t, 2, called1)
	assert.EqualValues(t, 1, called2)
}

func TestNewClient(t *testing.T) {
	c, err := NewClient(NewContext(nil, nil, map[string]string{
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
	c, err := NewClient(nil)
	assert.Nil(t, err)
	assert.False(t, c.SkipSSLVerify)

	for _, value := range []string{"true", "1", "t"} {
		c, err = NewClient(NewContext(nil, nil, map[string]string{
			"http.sslverify": value,
		}))
		t.Logf("http.sslverify: %q", value)
		assert.Nil(t, err)
		assert.False(t, c.SkipSSLVerify)
	}

	for _, value := range []string{"false", "0", "f"} {
		c, err = NewClient(NewContext(nil, nil, map[string]string{
			"http.sslverify": value,
		}))
		t.Logf("http.sslverify: %q", value)
		assert.Nil(t, err)
		assert.True(t, c.SkipSSLVerify)
	}
}

func TestNewClientWithOSSSLVerify(t *testing.T) {
	c, err := NewClient(nil)
	assert.Nil(t, err)
	assert.False(t, c.SkipSSLVerify)

	for _, value := range []string{"false", "0", "f"} {
		c, err = NewClient(NewContext(nil, map[string]string{
			"GIT_SSL_NO_VERIFY": value,
		}, nil))
		t.Logf("GIT_SSL_NO_VERIFY: %q", value)
		assert.Nil(t, err)
		assert.False(t, c.SkipSSLVerify)
	}

	for _, value := range []string{"true", "1", "t"} {
		c, err = NewClient(NewContext(nil, map[string]string{
			"GIT_SSL_NO_VERIFY": value,
		}, nil))
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
		c, err := NewClient(NewContext(nil, nil, map[string]string{
			"lfs.url": test[0],
		}))
		require.Nil(t, err)

		req, err := c.NewRequest("POST", c.Endpoints.Endpoint("", ""), test[1], nil)
		require.Nil(t, err)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, test[2], req.URL.String(), fmt.Sprintf("endpoint: %s, suffix: %s, expected: %s", test[0], test[1], test[2]))
	}
}

func TestNewRequestWithBody(t *testing.T) {
	c, err := NewClient(NewContext(nil, nil, map[string]string{
		"lfs.url": "https://example.com",
	}))
	require.Nil(t, err)

	body := struct {
		Test string
	}{Test: "test"}
	req, err := c.NewRequest("POST", c.Endpoints.Endpoint("", ""), "body", body)
	require.Nil(t, err)

	assert.NotNil(t, req.Body)
	assert.Equal(t, "15", req.Header.Get("Content-Length"))
	assert.EqualValues(t, 15, req.ContentLength)
}

func TestMarshalToRequest(t *testing.T) {
	req, err := http.NewRequest("POST", "https://foo/bar", nil)
	require.Nil(t, err)

	assert.Nil(t, req.Body)
	assert.Equal(t, "", req.Header.Get("Content-Length"))
	assert.EqualValues(t, 0, req.ContentLength)

	body := struct {
		Test string
	}{Test: "test"}
	require.Nil(t, MarshalToRequest(req, body))

	assert.NotNil(t, req.Body)
	assert.Equal(t, "15", req.Header.Get("Content-Length"))
	assert.EqualValues(t, 15, req.ContentLength)
}
