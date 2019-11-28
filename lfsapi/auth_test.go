package lfsapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/creds"
	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authRequest struct {
	Test string
}

func TestAuthenticateHeaderAccess(t *testing.T) {
	tests := map[string]creds.AccessMode{
		"":                creds.BasicAccess,
		"basic 123":       creds.BasicAccess,
		"basic":           creds.BasicAccess,
		"unknown":         creds.BasicAccess,
		"NTLM":            creds.NTLMAccess,
		"ntlm":            creds.NTLMAccess,
		"NTLM 1 2 3":      creds.NTLMAccess,
		"ntlm 1 2 3":      creds.NTLMAccess,
		"NEGOTIATE":       creds.NegotiateAccess,
		"negotiate":       creds.NegotiateAccess,
		"NEGOTIATE 1 2 3": creds.NegotiateAccess,
		"negotiate 1 2 3": creds.NegotiateAccess,
	}

	for _, key := range authenticateHeaders {
		for value, expected := range tests {
			res := &http.Response{Header: make(http.Header)}
			res.Header.Set(key, value)
			t.Logf("%s: %s", key, value)
			assert.Equal(t, expected, getAuthAccess(res))
		}
	}
}

func TestDoWithAuthApprove(t *testing.T) {
	var called uint32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&called, 1)
		assert.Equal(t, "POST", req.Method)

		body := &authRequest{}
		err := json.NewDecoder(req.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Approve", body.Test)

		w.Header().Set("Lfs-Authenticate", "Basic")
		actual := req.Header.Get("Authorization")
		if len(actual) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expected := "Basic " + strings.TrimSpace(
			base64.StdEncoding.EncodeToString([]byte("user:pass")),
		)
		assert.Equal(t, expected, actual)
	}))
	defer srv.Close()

	cred := newMockCredentialHelper()
	c, err := NewClient(lfshttp.NewContext(git.NewReadOnlyConfig("", ""),
		nil, map[string]string{
			"lfs.url": srv.URL + "/repo/lfs",
		},
	))
	require.Nil(t, err)
	c.Credentials = cred

	access := c.Endpoints.AccessFor(srv.URL + "/repo/lfs")
	assert.Equal(t, creds.NoneAccess, (&access).Mode())

	req, err := http.NewRequest("POST", srv.URL+"/repo/lfs/foo", nil)
	require.Nil(t, err)

	err = MarshalToRequest(req, &authRequest{Test: "Approve"})
	require.Nil(t, err)

	res, err := c.DoWithAuth("", c.Endpoints.AccessFor(srv.URL+"/repo/lfs"), req)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, cred.IsApproved(creds.Creds(map[string]string{
		"username": "user",
		"password": "pass",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})))
	access = c.Endpoints.AccessFor(srv.URL + "/repo/lfs")
	assert.Equal(t, creds.BasicAccess, (&access).Mode())
	assert.EqualValues(t, 2, called)
}

func TestDoWithAuthReject(t *testing.T) {
	var called uint32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&called, 1)
		assert.Equal(t, "POST", req.Method)

		body := &authRequest{}
		err := json.NewDecoder(req.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Reject", body.Test)

		actual := req.Header.Get("Authorization")
		expected := "Basic " + strings.TrimSpace(
			base64.StdEncoding.EncodeToString([]byte("user:pass")),
		)

		w.Header().Set("Lfs-Authenticate", "Basic")
		if actual != expected {
			// Write http.StatusUnauthorized to force the credential
			// helper to reject the credentials
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	invalidCreds := creds.Creds(map[string]string{
		"username": "user",
		"password": "wrong_pass",
		"path":     "",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})

	cred := newMockCredentialHelper()
	cred.Approve(invalidCreds)
	assert.True(t, cred.IsApproved(invalidCreds))

	c, _ := NewClient(nil)
	c.Credentials = cred
	c.Endpoints = NewEndpointFinder(lfshttp.NewContext(git.NewReadOnlyConfig("", ""),
		nil, map[string]string{
			"lfs.url": srv.URL,
		},
	))

	req, err := http.NewRequest("POST", srv.URL, nil)
	require.Nil(t, err)

	err = MarshalToRequest(req, &authRequest{Test: "Reject"})
	require.Nil(t, err)

	res, err := c.DoWithAuth("", c.Endpoints.AccessFor(srv.URL), req)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.False(t, cred.IsApproved(invalidCreds))
	assert.True(t, cred.IsApproved(creds.Creds(map[string]string{
		"username": "user",
		"password": "pass",
		"path":     "",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})))
	assert.EqualValues(t, 3, called)
}

func TestDoWithAuthNoRetry(t *testing.T) {
	var called uint32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&called, 1)
		assert.Equal(t, "POST", req.Method)

		body := &authRequest{}
		err := json.NewDecoder(req.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Approve", body.Test)

		w.Header().Set("Lfs-Authenticate", "Basic")
		actual := req.Header.Get("Authorization")
		if len(actual) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expected := "Basic " + strings.TrimSpace(
			base64.StdEncoding.EncodeToString([]byte("user:pass")),
		)
		assert.Equal(t, expected, actual)
	}))
	defer srv.Close()

	cred := newMockCredentialHelper()
	c, err := NewClient(lfshttp.NewContext(git.NewReadOnlyConfig("", ""),
		nil, map[string]string{
			"lfs.url": srv.URL + "/repo/lfs",
		},
	))
	require.Nil(t, err)
	c.Credentials = cred

	access := c.Endpoints.AccessFor(srv.URL + "/repo/lfs")
	assert.Equal(t, creds.NoneAccess, (&access).Mode())

	req, err := http.NewRequest("POST", srv.URL+"/repo/lfs/foo", nil)
	require.Nil(t, err)

	err = MarshalToRequest(req, &authRequest{Test: "Approve"})
	require.Nil(t, err)

	res, err := c.DoWithAuthNoRetry("", c.Endpoints.AccessFor(srv.URL+"/repo/lfs"), req)
	access = c.Endpoints.AccessFor(srv.URL + "/repo/lfs")
	assert.True(t, errors.IsAuthError(err))
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.Equal(t, creds.BasicAccess, (&access).Mode())
	assert.EqualValues(t, 1, called)
}

func TestDoAPIRequestWithAuth(t *testing.T) {
	var called uint32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddUint32(&called, 1)
		assert.Equal(t, "POST", req.Method)

		body := &authRequest{}
		err := json.NewDecoder(req.Body).Decode(body)
		assert.Nil(t, err)
		assert.Equal(t, "Approve", body.Test)

		w.Header().Set("Lfs-Authenticate", "Basic")
		actual := req.Header.Get("Authorization")
		if len(actual) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expected := "Basic " + strings.TrimSpace(
			base64.StdEncoding.EncodeToString([]byte("user:pass")),
		)
		assert.Equal(t, expected, actual)
	}))
	defer srv.Close()

	cred := newMockCredentialHelper()
	c, err := NewClient(lfshttp.NewContext(git.NewReadOnlyConfig("", ""),
		nil, map[string]string{
			"lfs.url": srv.URL + "/repo/lfs",
		},
	))
	require.Nil(t, err)
	c.Credentials = cred

	access := c.Endpoints.AccessFor(srv.URL + "/repo/lfs")
	assert.Equal(t, creds.NoneAccess, (&access).Mode())

	req, err := http.NewRequest("POST", srv.URL+"/repo/lfs/foo", nil)
	require.Nil(t, err)

	err = MarshalToRequest(req, &authRequest{Test: "Approve"})
	require.Nil(t, err)

	res, err := c.DoAPIRequestWithAuth("", req)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, cred.IsApproved(creds.Creds(map[string]string{
		"username": "user",
		"password": "pass",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})))
	access = c.Endpoints.AccessFor(srv.URL + "/repo/lfs")
	assert.Equal(t, creds.BasicAccess, (&access).Mode())
	assert.EqualValues(t, 2, called)
}

type mockCredentialHelper struct {
	Approved map[string]creds.Creds
}

func newMockCredentialHelper() *mockCredentialHelper {
	return &mockCredentialHelper{
		Approved: make(map[string]creds.Creds),
	}
}

func (m *mockCredentialHelper) Fill(input creds.Creds) (creds.Creds, error) {
	if found, ok := m.Approved[credsToKey(input)]; ok {
		return found, nil
	}

	output := make(creds.Creds)
	for key, value := range input {
		output[key] = value
	}
	if _, ok := output["username"]; !ok {
		output["username"] = "user"
	}
	output["password"] = "pass"
	return output, nil
}

func (m *mockCredentialHelper) Approve(creds creds.Creds) error {
	m.Approved[credsToKey(creds)] = creds
	return nil
}

func (m *mockCredentialHelper) Reject(creds creds.Creds) error {
	delete(m.Approved, credsToKey(creds))
	return nil
}

func (m *mockCredentialHelper) IsApproved(creds creds.Creds) bool {
	if found, ok := m.Approved[credsToKey(creds)]; ok {
		return found["password"] == creds["password"]
	}
	return false
}

func credsToKey(creds creds.Creds) string {
	var kvs []string
	for _, k := range []string{"protocol", "host", "path"} {
		kvs = append(kvs, fmt.Sprintf("%s:%s", k, creds[k]))
	}

	return strings.Join(kvs, " ")
}

func basicAuth(user, pass string) string {
	value := fmt.Sprintf("%s:%s", user, pass)
	return fmt.Sprintf("Basic %s", strings.TrimSpace(base64.StdEncoding.EncodeToString([]byte(value))))
}

type getCredsExpected struct {
	Access        creds.AccessMode
	Creds         creds.Creds
	CredsURL      string
	Authorization string
}

type getCredsTest struct {
	Remote   string
	Method   string
	Href     string
	Endpoint string
	Header   map[string]string
	Config   map[string]string
	Expected getCredsExpected
}

func TestGetCreds(t *testing.T) {
	tests := map[string]getCredsTest{
		"no access": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
			},
			Expected: getCredsExpected{
				Access: creds.NoneAccess,
			},
		},
		"basic access": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("git-server.com", "monkey"),
				CredsURL:      "https://git-server.com/repo/lfs",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com",
					"username": "git-server.com",
					"password": "monkey",
				},
			},
		},
		"basic access with usehttppath": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
				"credential.usehttppath":                     "true",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("git-server.com", "monkey"),
				CredsURL:      "https://git-server.com/repo/lfs",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com",
					"username": "git-server.com",
					"password": "monkey",
					"path":     "repo/lfs",
				},
			},
		},
		"basic access with url-specific usehttppath": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access":    "basic",
				"credential.https://git-server.com.usehttppath": "true",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("git-server.com", "monkey"),
				CredsURL:      "https://git-server.com/repo/lfs",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com",
					"username": "git-server.com",
					"password": "monkey",
					"path":     "repo/lfs",
				},
			},
		},
		"ntlm": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "ntlm",
			},
			Expected: getCredsExpected{
				Access:   creds.NTLMAccess,
				CredsURL: "https://git-server.com/repo/lfs",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com",
					"username": "git-server.com",
					"password": "monkey",
				},
			},
		},
		"custom auth": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Header: map[string]string{
				"Authorization": "custom",
			},
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: "custom",
			},
		},
		"username in url": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://user@git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("user", "monkey"),
				CredsURL:      "https://user@git-server.com/repo/lfs",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com",
					"username": "user",
					"password": "monkey",
				},
			},
		},
		"different remote url, basic access": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
				"remote.origin.url":                          "https://git-server.com/repo",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("git-server.com", "monkey"),
				CredsURL:      "https://git-server.com/repo",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com",
					"username": "git-server.com",
					"password": "monkey",
				},
			},
		},
		"api url auth": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/locks",
			Endpoint: "https://git-server.com/repo",
			Config: map[string]string{
				"lfs.url":                                "https://user:pass@git-server.com/repo",
				"lfs.https://git-server.com/repo.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("user", "pass"),
			},
		},
		"git url auth": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com/repo/locks",
			Endpoint: "https://git-server.com/repo",
			Config: map[string]string{
				"lfs.url":                                "https://git-server.com/repo",
				"lfs.https://git-server.com/repo.access": "basic",
				"remote.origin.url":                      "https://user:pass@git-server.com/repo",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("user", "pass"),
			},
		},
		"scheme mismatch": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "http://git-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("git-server.com", "monkey"),
				CredsURL:      "http://git-server.com/repo/lfs/locks",
				Creds: map[string]string{
					"protocol": "http",
					"host":     "git-server.com",
					"username": "git-server.com",
					"password": "monkey",
				},
			},
		},
		"host mismatch": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://lfs-server.com/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("lfs-server.com", "monkey"),
				CredsURL:      "https://lfs-server.com/repo/lfs/locks",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "lfs-server.com",
					"username": "lfs-server.com",
					"password": "monkey",
				},
			},
		},
		"port mismatch": getCredsTest{
			Remote:   "origin",
			Method:   "GET",
			Href:     "https://git-server.com:8080/repo/lfs/locks",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("git-server.com:8080", "monkey"),
				CredsURL:      "https://git-server.com:8080/repo/lfs/locks",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com:8080",
					"username": "git-server.com:8080",
					"password": "monkey",
				},
			},
		},
		"bare ssh URI": getCredsTest{
			Remote:   "origin",
			Method:   "POST",
			Href:     "https://git-server.com/repo/lfs/objects/batch",
			Endpoint: "https://git-server.com/repo/lfs",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",

				"remote.origin.url": "git@git-server.com:repo.git",
			},
			Expected: getCredsExpected{
				Access:        creds.BasicAccess,
				Authorization: basicAuth("git-server.com", "monkey"),
				CredsURL:      "https://git-server.com/repo/lfs",
				Creds: map[string]string{
					"host":     "git-server.com",
					"password": "monkey",
					"protocol": "https",
					"username": "git-server.com",
				},
			},
		},
	}

	for desc, test := range tests {
		t.Log(desc)
		req, err := http.NewRequest(test.Method, test.Href, nil)
		if err != nil {
			t.Errorf("[%s] %s", desc, err)
			continue
		}

		for key, value := range test.Header {
			req.Header.Set(key, value)
		}

		ctx := lfshttp.NewContext(git.NewReadOnlyConfig("", ""), nil, test.Config)
		client, _ := NewClient(ctx)
		client.Credentials = &fakeCredentialFiller{}
		client.Endpoints = NewEndpointFinder(ctx)
		credWrapper, err := client.getCreds(test.Remote, client.Endpoints.AccessFor(test.Endpoint), req)
		if !assert.Nil(t, err) {
			continue
		}

		assert.Equal(t, test.Expected.Authorization, req.Header.Get("Authorization"), "authorization")

		if test.Expected.Creds != nil {
			if desc == "ntlm" {
				// For NTLM we initially try with no provided credentials to test SSPI and then prompt.  We want to test both sets.
				assert.Nil(t, credWrapper.Creds, "creds")
				credWrapper.FillCreds()
			}
			assert.EqualValues(t, test.Expected.Creds, credWrapper.Creds)
		} else {
			assert.Nil(t, credWrapper.Creds, "creds")
		}

		if len(test.Expected.CredsURL) > 0 {
			if assert.NotNil(t, credWrapper.Url, "credURL") {
				assert.Equal(t, test.Expected.CredsURL, credWrapper.Url.String(), "credURL")
			}
		} else {
			assert.Nil(t, credWrapper.Url)
		}
	}
}

type fakeCredentialFiller struct{}

func (f *fakeCredentialFiller) Fill(input creds.Creds) (creds.Creds, error) {
	output := make(creds.Creds)
	for key, value := range input {
		output[key] = value
	}
	if _, ok := output["username"]; !ok {
		output["username"] = input["host"]
	}
	output["password"] = "monkey"
	return output, nil
}

func (f *fakeCredentialFiller) Approve(creds creds.Creds) error {
	return errors.New("Not implemented")
}

func (f *fakeCredentialFiller) Reject(creds creds.Creds) error {
	return errors.New("Not implemented")
}

func TestClientRedirectReauthenticate(t *testing.T) {
	var srv1, srv2 *httptest.Server
	var called1, called2 uint32
	var creds1, creds2 creds.Creds

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

	creds1 = creds.Creds(map[string]string{
		"protocol": "http",
		"host":     strings.TrimPrefix(srv1.URL, "http://"),

		"username": "user1",
		"password": "pass1",
	})
	creds2 = creds.Creds(map[string]string{
		"protocol": "http",
		"host":     strings.TrimPrefix(srv2.URL, "http://"),

		"username": "user2",
		"password": "pass2",
	})

	defer srv1.Close()
	defer srv2.Close()

	c, err := NewClient(lfshttp.NewContext(nil, nil, nil))
	cred := creds.NewCredentialCacher()
	cred.Approve(creds1)
	cred.Approve(creds2)
	c.Credentials = cred

	req, err := http.NewRequest("GET", srv1.URL, nil)
	require.Nil(t, err)

	_, err = c.DoAPIRequestWithAuth("", req)
	assert.Nil(t, err)

	// called1 is 2 since LFS tries an unauthenticated request first
	assert.EqualValues(t, 2, called1)
	assert.EqualValues(t, 1, called2)
}
