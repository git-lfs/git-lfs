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

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authRequest struct {
	Test string
}

func TestAuthenticateHeaderAccess(t *testing.T) {
	tests := map[string]Access{
		"":                BasicAccess,
		"basic 123":       BasicAccess,
		"basic":           BasicAccess,
		"unknown":         BasicAccess,
		"NTLM":            NTLMAccess,
		"ntlm":            NTLMAccess,
		"NTLM 1 2 3":      NTLMAccess,
		"ntlm 1 2 3":      NTLMAccess,
		"NEGOTIATE":       NTLMAccess,
		"negotiate":       NTLMAccess,
		"NEGOTIATE 1 2 3": NTLMAccess,
		"negotiate 1 2 3": NTLMAccess,
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

	creds := newMockCredentialHelper()
	c, err := NewClient(NewContext(nil, nil, map[string]string{
		"lfs.url": srv.URL + "/repo/lfs",
	}))
	require.Nil(t, err)
	c.Credentials = creds

	assert.Equal(t, NoneAccess, c.Endpoints.AccessFor(srv.URL+"/repo/lfs"))

	req, err := http.NewRequest("POST", srv.URL+"/repo/lfs/foo", nil)
	require.Nil(t, err)

	err = MarshalToRequest(req, &authRequest{Test: "Approve"})
	require.Nil(t, err)

	res, err := c.DoWithAuth("", req)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, creds.IsApproved(Creds(map[string]string{
		"username": "user",
		"password": "pass",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})))
	assert.Equal(t, BasicAccess, c.Endpoints.AccessFor(srv.URL+"/repo/lfs"))
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

	invalidCreds := Creds(map[string]string{
		"username": "user",
		"password": "wrong_pass",
		"path":     "",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})

	creds := newMockCredentialHelper()
	creds.Approve(invalidCreds)
	assert.True(t, creds.IsApproved(invalidCreds))

	c, _ := NewClient(nil)
	c.Credentials = creds
	c.Endpoints = NewEndpointFinder(NewContext(nil, nil, map[string]string{
		"lfs.url": srv.URL,
	}))

	req, err := http.NewRequest("POST", srv.URL, nil)
	require.Nil(t, err)

	err = MarshalToRequest(req, &authRequest{Test: "Reject"})
	require.Nil(t, err)

	res, err := c.DoWithAuth("", req)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.False(t, creds.IsApproved(invalidCreds))
	assert.True(t, creds.IsApproved(Creds(map[string]string{
		"username": "user",
		"password": "pass",
		"path":     "",
		"protocol": "http",
		"host":     srv.Listener.Addr().String(),
	})))
	assert.EqualValues(t, 3, called)
}

type mockCredentialHelper struct {
	Approved map[string]Creds
}

func newMockCredentialHelper() *mockCredentialHelper {
	return &mockCredentialHelper{
		Approved: make(map[string]Creds),
	}
}

func (m *mockCredentialHelper) Fill(input Creds) (Creds, error) {
	if found, ok := m.Approved[credsToKey(input)]; ok {
		return found, nil
	}

	output := make(Creds)
	for key, value := range input {
		output[key] = value
	}
	if _, ok := output["username"]; !ok {
		output["username"] = "user"
	}
	output["password"] = "pass"
	return output, nil
}

func (m *mockCredentialHelper) Approve(creds Creds) error {
	m.Approved[credsToKey(creds)] = creds
	return nil
}

func (m *mockCredentialHelper) Reject(creds Creds) error {
	delete(m.Approved, credsToKey(creds))
	return nil
}

func (m *mockCredentialHelper) IsApproved(creds Creds) bool {
	if found, ok := m.Approved[credsToKey(creds)]; ok {
		return found["password"] == creds["password"]
	}
	return false
}

func credsToKey(creds Creds) string {
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
	Endpoint      string
	Access        Access
	Creds         Creds
	CredsURL      string
	Authorization string
}

type getCredsTest struct {
	Remote   string
	Method   string
	Href     string
	Header   map[string]string
	Config   map[string]string
	Expected getCredsExpected
}

func TestGetCreds(t *testing.T) {
	tests := map[string]getCredsTest{
		"no access": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
			},
			Expected: getCredsExpected{
				Access:   NoneAccess,
				Endpoint: "https://git-server.com/repo/lfs",
			},
		},
		"basic access": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
				"credential.usehttppath":                     "true",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access":    "basic",
				"credential.https://git-server.com.usehttppath": "true",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "ntlm",
			},
			Expected: getCredsExpected{
				Access:   NTLMAccess,
				Endpoint: "https://git-server.com/repo/lfs",
				CredsURL: "https://git-server.com/repo/lfs",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "git-server.com",
					"username": "git-server.com",
					"password": "monkey",
				},
			},
		},
		"ntlm with netrc": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "https://netrc-host.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://netrc-host.com/repo/lfs",
				"lfs.https://netrc-host.com/repo/lfs.access": "ntlm",
			},
			Expected: getCredsExpected{
				Access:   NTLMAccess,
				Endpoint: "https://netrc-host.com/repo/lfs",
				CredsURL: "https://netrc-host.com/repo/lfs",
				Creds: map[string]string{
					"protocol": "https",
					"host":     "netrc-host.com",
					"username": "abc",
					"password": "def",
					"source":   "netrc",
				},
			},
		},
		"custom auth": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Header: map[string]string{
				"Authorization": "custom",
			},
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
				Authorization: "custom",
			},
		},
		"netrc": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "https://netrc-host.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://netrc-host.com/repo/lfs",
				"lfs.https://netrc-host.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://netrc-host.com/repo/lfs",
				Authorization: basicAuth("abc", "def"),
			},
		},
		"username in url": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://user@git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://user@git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
				"remote.origin.url":                          "https://git-server.com/repo",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/locks",
			Config: map[string]string{
				"lfs.url": "https://user:pass@git-server.com/repo",
				"lfs.https://git-server.com/repo.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://user:pass@git-server.com/repo",
				Authorization: basicAuth("user", "pass"),
			},
		},
		"git url auth": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com/repo/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo",
				"lfs.https://git-server.com/repo.access": "basic",
				"remote.origin.url":                      "https://user:pass@git-server.com/repo",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo",
				Authorization: basicAuth("user", "pass"),
			},
		},
		"scheme mismatch": getCredsTest{
			Remote: "origin",
			Method: "GET",
			Href:   "http://git-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "GET",
			Href:   "https://lfs-server.com/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "GET",
			Href:   "https://git-server.com:8080/repo/lfs/locks",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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
			Remote: "origin",
			Method: "POST",
			Href:   "https://git-server.com/repo/lfs/objects/batch",
			Config: map[string]string{
				"lfs.url": "https://git-server.com/repo/lfs",
				"lfs.https://git-server.com/repo/lfs.access": "basic",

				"remote.origin.url": "git@git-server.com:repo.git",
			},
			Expected: getCredsExpected{
				Access:        BasicAccess,
				Endpoint:      "https://git-server.com/repo/lfs",
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

		ctx := NewContext(git.NewConfig("", ""), nil, test.Config)
		client, _ := NewClient(ctx)
		client.Credentials = &fakeCredentialFiller{}
		client.Netrc = &fakeNetrc{}
		client.Endpoints = NewEndpointFinder(ctx)
		endpoint, access, _, credsURL, creds, err := client.getCreds(test.Remote, req)
		if !assert.Nil(t, err) {
			continue
		}
		assert.Equal(t, test.Expected.Endpoint, endpoint.Url, "endpoint")
		assert.Equal(t, test.Expected.Access, access, "access")
		assert.Equal(t, test.Expected.Authorization, req.Header.Get("Authorization"), "authorization")

		if test.Expected.Creds != nil {
			assert.EqualValues(t, test.Expected.Creds, creds)
		} else {
			assert.Nil(t, creds, "creds")
		}

		if len(test.Expected.CredsURL) > 0 {
			if assert.NotNil(t, credsURL, "credURL") {
				assert.Equal(t, test.Expected.CredsURL, credsURL.String(), "credURL")
			}
		} else {
			assert.Nil(t, credsURL)
		}
	}
}

type fakeCredentialFiller struct{}

func (f *fakeCredentialFiller) Fill(input Creds) (Creds, error) {
	output := make(Creds)
	for key, value := range input {
		output[key] = value
	}
	if _, ok := output["username"]; !ok {
		output["username"] = input["host"]
	}
	output["password"] = "monkey"
	return output, nil
}

func (f *fakeCredentialFiller) Approve(creds Creds) error {
	return errors.New("Not implemented")
}

func (f *fakeCredentialFiller) Reject(creds Creds) error {
	return errors.New("Not implemented")
}
