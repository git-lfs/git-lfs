package lfsapi

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/git-lfs/git-lfs/creds"
	"github.com/git-lfs/git-lfs/lfshttp"
	"github.com/git-lfs/go-ntlm/ntlm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNtlmAuth(t *testing.T) {
	session, err := ntlm.CreateServerSession(ntlm.Version2, ntlm.ConnectionOrientedMode)
	require.Nil(t, err)
	session.SetUserInfo("ntlmuser", "ntlmpass", "NTLMDOMAIN")

	var called uint32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reqIndex := atomic.LoadUint32(&called)
		atomic.AddUint32(&called, 1)

		if called == 4 && runtime.GOOS != "windows" {
			// We don't want to hit the second class 1 path if we are on *nix.
			atomic.AddUint32(&called, 1)
		}

		authHeader := req.Header.Get("Authorization")
		t.Logf("REQUEST %d: %s %s", reqIndex, req.Method, req.URL)
		t.Logf("AUTH: %q", authHeader)

		// assert full body is sent each time
		by, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if assert.Nil(t, err) {
			assert.Equal(t, "ntlm", string(by))
		}

		switch called {
		case 1:
			w.Header().Set("Www-Authenticate", "ntlm")
			w.WriteHeader(401)
		case 2, 4:
			assert.True(t, strings.HasPrefix(req.Header.Get("Authorization"), "NTLM "))
			neg := authHeader[5:] // strip "ntlm " prefix
			_, err := base64.StdEncoding.DecodeString(neg)
			if !assert.Nil(t, err) {
				t.Logf("neg base64 error: %+v", err)
				w.WriteHeader(500)
				return
			}

			// ntlm implementation can't currently parse the negotiate message
			ch, err := session.GenerateChallengeMessage()
			if !assert.Nil(t, err) {
				t.Logf("challenge gen error: %+v", err)
				w.WriteHeader(500)
				return
			}
			chMsg := base64.StdEncoding.EncodeToString(ch.Bytes())
			w.Header().Set("Www-Authenticate", "ntlm "+chMsg)
			w.WriteHeader(401)
		default: // should be an auth msg
			authHeader := req.Header.Get("Authorization")
			assert.True(t, strings.HasPrefix(strings.ToUpper(authHeader), "NTLM "))
			auth := authHeader[5:] // strip "ntlm " prefix
			val, err := base64.StdEncoding.DecodeString(auth)
			if !assert.Nil(t, err) {
				t.Logf("auth base64 error: %+v", err)
				w.WriteHeader(500)
				return
			}

			authMsg, err := ntlm.ParseAuthenticateMessage(val, 2)
			if !assert.Nil(t, err) {
				t.Logf("auth parse error: %+v", err)
				w.WriteHeader(500)
				return
			}

			if called == 3 && runtime.GOOS == "windows" {
				// This is the SSPI call that should return unauth so that standard NTLM can run.
				w.WriteHeader(401)
				return
			}

			err = session.ProcessAuthenticateMessage(authMsg)
			if !assert.Nil(t, err) {
				t.Logf("auth process error: %+v", err)
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)

		}
	}))
	defer srv.Close()

	req, err := http.NewRequest("POST", srv.URL+"/ntlm", NewByteBody([]byte("ntlm")))
	require.Nil(t, err)

	credHelper := newMockCredentialHelper()
	cli, err := NewClient(lfshttp.NewContext(nil, nil, map[string]string{
		"lfs.url":                         srv.URL + "/ntlm",
		"lfs." + srv.URL + "/ntlm.access": "ntlm",
	}))
	cli.Credentials = credHelper
	require.Nil(t, err)

	// ntlm support pulls domain and login info from git credentials
	srvURL, err := url.Parse(srv.URL)
	require.Nil(t, err)
	cred := creds.Creds{
		"protocol": srvURL.Scheme,
		"host":     srvURL.Host,
		"username": "ntlmdomain\\ntlmuser",
		"password": "ntlmpass",
	}
	credHelper.Approve(cred)

	res, err := cli.DoWithAuth("remote", cli.Endpoints.AccessFor(srv.URL+"/ntlm"), req)
	require.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.True(t, credHelper.IsApproved(cred))
}

func TestNtlmGetCredentials(t *testing.T) {
	cred := creds.Creds{"username": "MOOSEDOMAIN\\canadian", "password": "MooseAntlersYeah"}
	ntmlCreds, err := ntlmGetCredentials(cred)
	assert.Nil(t, err)
	assert.NotNil(t, ntmlCreds)
	assert.Equal(t, "MOOSEDOMAIN", ntmlCreds.domain)
	assert.Equal(t, "canadian", ntmlCreds.username)
	assert.Equal(t, "MooseAntlersYeah", ntmlCreds.password)

	cred = creds.Creds{"username": "", "password": ""}
	ntmlCreds, err = ntlmGetCredentials(cred)
	assert.Nil(t, err)
	assert.Nil(t, ntmlCreds)
}

func TestNtlmGetCredentialsBadCreds(t *testing.T) {
	cred := creds.Creds{"username": "badusername", "password": "MooseAntlersYeah"}
	_, err := ntlmGetCredentials(cred)
	assert.NotNil(t, err)
}

func TestNtlmHeaderParseValid(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", "NTLM "+base64.StdEncoding.EncodeToString([]byte("I am a moose")))
	bytes, err := parseChallengeResponse(&res)
	assert.Nil(t, err)
	assert.False(t, strings.HasPrefix(string(bytes), "NTLM"))
}

func TestNtlmHeaderParseInvalidLength(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", "NTL")
	ret, err := parseChallengeResponse(&res)
	assert.NotNil(t, err)
	assert.Nil(t, ret)
}

func TestNtlmHeaderParseInvalid(t *testing.T) {
	res := http.Response{}
	res.Header = make(map[string][]string)
	res.Header.Add("Www-Authenticate", base64.StdEncoding.EncodeToString([]byte("NTLM I am a moose")))
	ret, err := parseChallengeResponse(&res)
	assert.NotNil(t, err)
	assert.Nil(t, ret)
}

func assertRequestsEqual(t *testing.T, req1 *http.Request, req2 *http.Request, req1Body []byte) {
	assert.Equal(t, req1.Method, req2.Method)

	for k, v := range req1.Header {
		assert.Equal(t, v, req2.Header[k])
	}

	if req1.Body == nil {
		assert.Nil(t, req2.Body)
	} else {
		bytes2, _ := ioutil.ReadAll(req2.Body)
		assert.Equal(t, req1Body, bytes2)
	}
}
